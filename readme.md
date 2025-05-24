# golib

A modular Go library for structured logging, Prometheus metrics, PostgreSQL transaction handling, and health checks. Designed for production services, testability, and clean integration with `context.Context`.

---

## Features

- ✅ **Structured Logging with Context** via `slog`
- 📊 **Dynamic Prometheus Metrics** (count + duration)
- 🔍 **Health Probes** for Kubernetes: `/healthz` and `/readyz`
- 🧾 **PostgreSQL Transaction Manager** with context propagation
- 🔌 **Stub Implementations** for tests

---

## 🧱 Packages

### `clog` – Context-Aware Logging

Structured JSON logging using `slog`, with context support and error separation.

```go
logger := clog.NewCustomLogger(os.Stdout, os.Stderr, true, slog.LevelInfo)
logger.InfoCtx(ctx, "user created: %s", userID)
logger.ErrorCtx(ctx, err, "failed to create user")
```

---

### `metrics` – Prometheus Metrics Layer

Track count and duration metrics per series type and operation. Automatically registers dynamic metrics with labels.

#### Series

```go
series := metrics.NewSeries(metrics.SeriesTypeAPIHandler, "v1").
    WithLabels(prometheus.Labels{"region": "us-east"})
ctx, series = series.WithOperation(ctx, "getUser")
```

#### Registry

```go
reg := metrics.NewRegistry("api", "myapp")

name, labels := series.Success()
reg.Inc(name, labels)

name, labels, duration := series.Duration(time.Since(start))
reg.RecordDuration(name, labels, duration)
```

#### Server + Health

```go
health := metrics.NewHealthChecker(logger)
server := metrics.NewServer(logger, reg, health, ":9090", 5*time.Second)

go server.Start(context.Background())
```

Exposes:
- `GET /metrics`
- `GET /healthz`
- `GET /readyz`

---

### `pgv10` – PostgreSQL Transaction Management

Wraps `go-pg/pg/v10`, propagates transactions via context, supports automatic rollback for errors.

#### Usage

```go
factory := pgv10.NewPgTransactionFactory(db)
manager := pgv10.NewPgTransactionManager(factory, pgv10.Options{AlwaysRollback: false})

err := manager.Do(ctx, func(ctx context.Context) error {
	tx := factory.Transaction(ctx)
	// run queries with tx
	return nil
})
```

#### Test Stub

```go
stub := pgv10.NewTrmStub()
_ = stub.Do(ctx, func(ctx context.Context) error {
	// test logic
	return nil
})
```

---

### `metrics.RegistryStub`

For disabling metric collection in tests.

```go
reg := metrics.NewRegistryStub()
reg.Inc("noop_metric", nil)
```

---

## ✍️ License

MIT

---

## 🚀 Example Integration

```go
reg := metrics.NewRegistry("api", "neonrpc")
logger := clog.NewCustomLogger(os.Stdout, os.Stderr, false, slog.LevelDebug)
health := metrics.NewHealthChecker(logger)
server := metrics.NewServer(logger, reg, health, ":8081", 5*time.Second)
go server.Start(ctx)

series := metrics.NewSeries(metrics.SeriesTypeClient, "uscis")
ctx, series = series.WithOperation(ctx, "checkStatus")
name, labels := series.Success()
reg.Inc(name, labels)
```

```go
pgTxFactory := pgv10.NewPgTransactionFactory(pgdb)
pgTxManager := pgv10.NewPgTransactionManager(pgTxFactory, pgv10.Options{AlwaysRollback: false})
_ = pgTxManager.Do(ctx, func(ctx context.Context) error {
	// logic here
	return nil
})
```
