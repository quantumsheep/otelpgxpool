package otelpgxpool

import (
	"github.com/exaring/otelpgx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// Option specifies instrumentation configuration options.
type Option interface {
	apply(*tracerConfig)
}

type optionFunc func(*tracerConfig)

func (o optionFunc) apply(c *tracerConfig) {
	o(c)
}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return optionFunc(func(cfg *tracerConfig) {
		if provider != nil {
			cfg.tp = provider
		}
	})
}

// WithAttributes specifies additional attributes to be added to the span.
func WithAttributes(attrs ...attribute.KeyValue) Option {
	return optionFunc(func(cfg *tracerConfig) {
		cfg.attrs = append(cfg.attrs, attrs...)
	})
}

// WithOtelPgxTracer specifies an otelpgx.Tracer to use for tracing pgx queries.
func WithOtelPgxTracer(tracer *otelpgx.Tracer) Option {
	return optionFunc(func(cfg *tracerConfig) {
		cfg.otelpgxTracer = tracer
	})
}
