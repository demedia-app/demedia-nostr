package trace

import (
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type TracerConfig struct {
	TraceExporter  string
	ServiceName    string
	Environment    string
	ServiceVersion string
}

func CreateTracers(cfg TracerConfig) (trace.Tracer, func(context.Context) error) {
	exp, err := createOtelExporter(cfg.TraceExporter)
	if err != nil {
		log.Fatalf("cannot create trace exporter %s: %w", cfg.TraceExporter, err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
			semconv.DeploymentEnvironmentKey.String(cfg.Environment),
		)),
	)

	return tp.Tracer(cfg.ServiceName), func(ctx context.Context) error {
		if err := tp.Shutdown(ctx); err != nil {
			return err
		}

		return nil
	}
}

func createOtelExporter(exporterType string) (sdktrace.SpanExporter, error) {
	var exporter sdktrace.SpanExporter
	var err error
	switch exporterType {
	case "jaeger":
		exporter, err = jaeger.New(
			jaeger.WithCollectorEndpoint(),
		)
	case "otlp", "otlp-http":
		client := otlptracehttp.NewClient(
			otlptracehttp.WithInsecure(),
		)
		exporter, err = otlptrace.New(context.Background(), client)
	case "otlp-grpc":
		client := otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
		)
		exporter, err = otlptrace.New(context.Background(), client)
	case "stdout":
		exporter, err = stdouttrace.New()
	default:
		return nil, fmt.Errorf("unrecognized exporter type %s", exporterType)
	}
	return exporter, err
}
