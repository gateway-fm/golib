package clog_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/yusupovanton/golib/clog"
)

func BenchmarkClog(b *testing.B) {
	logger := clog.NewCustomLogger(io.Discard, io.Discard, false, slog.LevelInfo)
	ctx := context.WithValue(b.Context(), "userID", 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoCtx(ctx, "benchmarking clog")
	}
}

func BenchmarkSlog(b *testing.B) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelInfo}))
	ctx := context.WithValue(b.Context(), "userID", 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoContext(ctx, "benchmarking slog", slog.Int("userID", 12345))
	}
}

func BenchmarkZap(b *testing.B) {
	cfg := zap.NewProductionConfig()
	cfg.OutputPaths = []string{}
	cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	cfg.EncoderConfig.TimeKey = ""
	cfg.EncoderConfig.LevelKey = ""

	logger, _ := cfg.Build()
	defer func() {
		err := logger.Sync()
		require.NoError(b, err)
	}()
	sugar := logger.Sugar()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sugar.Infow("benchmarking zap", "userID", 12345)
	}
}
