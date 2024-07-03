[![Go Reference](https://pkg.go.dev/badge/github.com/quantumsheep/otelpgxpool.svg)](https://pkg.go.dev/github.com/quantumsheep/otelpgxpool)

# otelpgx

Provides [OpenTelemetry](https://github.com/open-telemetry/opentelemetry-go) instrumentation for the [jackc/pgx](https://github.com/jackc/pgx) library.

It has support for [otelpgx](https://github.com/exaring/otelpgx) which the library is based on.

## Requirements

- go 1.21 (or higher)
- pgx v5 (or higher)

## Usage

Install the library:

```go
go get github.com/quantumsheep/otelpgxpool
```

Create the tracer as part of your connection:

```go
cfg, err := pgxpool.ParseConfig(connString)
if err != nil {
    return nil, fmt.Errorf("create connection pool: %w", err)
}

cfg.ConnConfig.Tracer = otelpgx.NewTracer()

conn, err := pgxpool.NewWithConfig(ctx, cfg)
if err != nil {
    return nil, fmt.Errorf("connect to database: %w", err)
}
```

See [options.go](options.go) for the full list of options.
