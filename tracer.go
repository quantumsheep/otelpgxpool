package otelpgxpool

import (
	"context"
	"runtime/debug"

	"github.com/exaring/otelpgx"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "github.com/quantumsheep/otelpgxpool"

type OtelPgxTracer = otelpgx.Tracer

// Tracer is a wrapper around the pgx tracer interfaces which instrument
// queries.
type Tracer struct {
	*OtelPgxTracer
	tracer trace.Tracer
	attrs  []attribute.KeyValue
}

var (
	_ pgx.QueryTracer       = (*Tracer)(nil)
	_ pgx.BatchTracer       = (*Tracer)(nil)
	_ pgx.CopyFromTracer    = (*Tracer)(nil)
	_ pgx.PrepareTracer     = (*Tracer)(nil)
	_ pgx.ConnectTracer     = (*Tracer)(nil)
	_ pgxpool.AcquireTracer = (*Tracer)(nil)
)

type tracerConfig struct {
	otelpgxTracer *otelpgx.Tracer
	tp            trace.TracerProvider
	attrs         []attribute.KeyValue
}

// NewTracer returns a new Tracer.
func NewTracer(opts ...Option) *Tracer {
	cfg := &tracerConfig{
		tp: otel.GetTracerProvider(),
		attrs: []attribute.KeyValue{
			semconv.DBSystemPostgreSQL,
		},
	}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	if cfg.otelpgxTracer == nil {
		cfg.otelpgxTracer = otelpgx.NewTracer(
			otelpgx.WithTracerProvider(cfg.tp),
			otelpgx.WithAttributes(cfg.attrs...),
		)
	}

	return &Tracer{
		OtelPgxTracer: cfg.otelpgxTracer,
		tracer:        cfg.tp.Tracer(tracerName, trace.WithInstrumentationVersion(findOwnImportedVersion())),
		attrs:         cfg.attrs,
	}
}

func recordError(span trace.Span, err error) {
	if err == nil {
		return
	}

	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
}

// connectionAttributesFromConfig returns a slice of SpanStartOptions that contain
// attributes from the given connection config.
func connectionAttributesFromConfig(config *pgx.ConnConfig) []trace.SpanStartOption {
	if config != nil {
		return []trace.SpanStartOption{
			trace.WithAttributes(
				attribute.String("server.address", config.Host),
				attribute.Int("server.port", int(config.Port)),
				semconv.DBUser(config.User),
			),
		}
	}
	return nil
}

// TraceAcquireStart is called at the beginning of Acquire.
// The returned context is used for the rest of the call and will be passed to the TraceAcquireEnd.
func (t *Tracer) TraceAcquireStart(ctx context.Context, pool *pgxpool.Pool, data pgxpool.TraceAcquireStartData) context.Context {
	if !trace.SpanFromContext(ctx).IsRecording() {
		return ctx
	}

	opts := []trace.SpanStartOption{
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(t.attrs...),
	}

	if pool != nil {
		config := pool.Config()

		if config != nil && config.ConnConfig != nil {
			opts = append(opts, connectionAttributesFromConfig(pool.Config().ConnConfig)...)
		}
	}

	ctx, _ = t.tracer.Start(ctx, "acquire", opts...)

	return ctx
}

// TraceAcquireEnd is called when a connection has been acquired.
func (t *Tracer) TraceAcquireEnd(ctx context.Context, pool *pgxpool.Pool, data pgxpool.TraceAcquireEndData) {
	span := trace.SpanFromContext(ctx)
	recordError(span, data.Err)

	span.End()
}

func findOwnImportedVersion() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if ok {
		for _, dep := range buildInfo.Deps {
			if dep.Path == tracerName {
				return dep.Version
			}
		}
	}

	return "unknown"
}
