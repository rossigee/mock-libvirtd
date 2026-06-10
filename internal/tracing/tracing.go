package tracing

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"log/slog"
)

type TracerProvider = sdktrace.TracerProvider

var tracer *sdktrace.TracerProvider

func Init(serviceName string) *TracerProvider {
	endpoint := os.Getenv("OTLP_ENDPOINT")
	if endpoint == "" {
		return nil
	}

	exporter, err := otlptracegrpc.New(context.Background(),
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		slog.Warn("failed to create OTLP exporter", slog.String("endpoint", endpoint), slog.Any("error", err))
		return nil
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
		),
	)
	if err != nil {
		slog.Warn("failed to create OTLP resource", slog.Any("error", err))
		return nil
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer = tp
	slog.Info("OTLP tracing enabled", slog.String("endpoint", endpoint))
	return tp
}

func Shutdown() {
	if tracer != nil {
		_ = tracer.Shutdown(context.Background())
	}
}