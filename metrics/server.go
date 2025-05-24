package metrics

import (
	"context"
	"errors"
	"net/http"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/yusupovanton/golib/clog"
)

const (
	metricsEndpoint   = "/metrics"
	livenessEndpoint  = "/healthz"
	readinessEndpoint = "/readyz"
)

var (
	memAllocGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_mem_stats_alloc_bytes",
		Help: "Number of bytes allocated and still in use.",
	})
	memSysGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_mem_stats_sys_bytes",
		Help: "Number of bytes obtained from the system.",
	})
	memHeapAllocGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_mem_stats_heap_alloc_bytes",
		Help: "Number of heap bytes allocated and still in use.",
	})
	memHeapSysGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "go_mem_stats_heap_sys_bytes",
		Help: "Number of heap bytes obtained from the system.",
	})
)

type server struct {
	logger         clog.CLog
	registry       Registry
	healthCheck    *HealthChecker
	httpServer     *http.Server
	stopCh         chan struct{}
	addr           string
	scrapeInterval time.Duration
}

func NewServer(
	logger clog.CLog,
	registry Registry,
	healthCheck *HealthChecker,
	addr string,
	scrapeInterval time.Duration,
) Server {
	registry.PrometheusRegistry().MustRegister(memAllocGauge, memSysGauge, memHeapAllocGauge, memHeapSysGauge)

	return &server{
		logger:         logger,
		registry:       registry,
		healthCheck:    healthCheck,
		stopCh:         make(chan struct{}),
		addr:           addr,
		scrapeInterval: scrapeInterval,
	}
}

func (s *server) Start(ctx context.Context) {
	mux := http.NewServeMux()

	mux.Handle(metricsEndpoint, promhttp.HandlerFor(s.registry.PrometheusRegistry(), promhttp.HandlerOpts{}))
	mux.HandleFunc(livenessEndpoint, s.healthCheck.LivenessHandler)
	mux.HandleFunc(readinessEndpoint, s.healthCheck.ReadinessHandler)

	s.httpServer = &http.Server{
		Addr:              s.addr,
		Handler:           mux,
		ReadHeaderTimeout: 1 * time.Second,
		ReadTimeout:       2 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       30 * time.Second,
	}

	go s.collectMemoryStats(ctx)

	go func() {
		s.logger.InfoCtx(ctx, "starting metrics server, address: %s", s.addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.ErrorCtx(ctx, err, "failed to start metrics server")
		}
	}()
}

func (s *server) Stop(ctx context.Context) error {
	close(s.stopCh)

	if s.httpServer != nil {
		err := s.httpServer.Shutdown(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *server) collectMemoryStats(ctx context.Context) {
	ticker := time.NewTicker(s.scrapeInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.InfoCtx(ctx, "stopping memory stats collection")
			return
		case <-s.stopCh:
			s.logger.InfoCtx(ctx, "memory stats collection stopped by server stop")
			return
		case <-ticker.C:
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			memAllocGauge.Set(float64(memStats.Alloc))
			memSysGauge.Set(float64(memStats.Sys))
			memHeapAllocGauge.Set(float64(memStats.HeapAlloc))
			memHeapSysGauge.Set(float64(memStats.HeapSys))
		}
	}
}
