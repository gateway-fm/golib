package clog_test

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/yusupovanton/golib/clog"
)

func BenchmarkCustomLogger(b *testing.B) {
	var buf bytes.Buffer

	logger := clog.NewCustomLogger(&buf, &buf, false, slog.LevelDebug)

	ctx := logger.AddKeysValuesToCtx(b.Context(), map[string]interface{}{
		"userID":    12345,
		"userName":  "testuser",
		"timestamp": time.Now(),
		"data":      []int{1, 2, 3},
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.InfoCtx(ctx, "Some test message")
	}
}
