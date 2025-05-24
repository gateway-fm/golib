package metrics

import (
	"sort"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type registry struct {
	Subsystem    string
	Namespace    string
	PromRegistry *prometheus.Registry

	metricsMu  sync.Mutex
	counters   map[string]*prometheus.CounterVec
	histograms map[string]*prometheus.HistogramVec
}

func NewRegistry(subsystem, namespace string) Registry {
	r := &registry{
		Subsystem:    subsystem,
		Namespace:    namespace,
		PromRegistry: prometheus.NewRegistry(),
		counters:     make(map[string]*prometheus.CounterVec),
		histograms:   make(map[string]*prometheus.HistogramVec),
	}

	registerMetrics(r)

	return r
}

// Inc increments a counter for the given Series, dynamically determining label names from the input labels.
func (r *registry) Inc(name string, labels prometheus.Labels) {
	r.metricsMu.Lock()
	defer r.metricsMu.Unlock()

	counter, exists := r.counters[name]
	if !exists {
		counter = prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: r.Subsystem,
			Namespace: r.Namespace,
			Name:      name,
		}, getLabelNames(labels))
		r.PromRegistry.MustRegister(counter)
		r.counters[name] = counter
	}
	counter.With(labels).Inc()
}

// RecordDuration records a duration for the given Series, dynamically determining label names from the input labels.
func (r *registry) RecordDuration(name string, labels prometheus.Labels, duration float64) {
	r.metricsMu.Lock()
	defer r.metricsMu.Unlock()

	histogram, exists := r.histograms[name]
	if !exists {
		histogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: r.Subsystem,
			Namespace: r.Namespace,
			Name:      name,
			Buckets:   prometheus.DefBuckets,
		}, getLabelNames(labels))
		r.PromRegistry.MustRegister(histogram)
		r.histograms[name] = histogram
	}
	histogram.With(labels).Observe(duration)
}

// PrometheusRegistry returns the underlying Prometheus registry.
func (r *registry) PrometheusRegistry() *prometheus.Registry {
	return r.PromRegistry
}

func registerMetrics(registry *registry) {
	registry.PromRegistry.MustRegister(
		collectors.NewGoCollector(
			collectors.WithGoCollectorMemStatsMetricsDisabled(),
			collectors.WithGoCollectorRuntimeMetrics(collectors.MetricsScheduler),
		))
}

// getLabelNames returns the keys of the labels map as a slice of strings, sorted to ensure consistent order.
func getLabelNames(labels prometheus.Labels) []string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
