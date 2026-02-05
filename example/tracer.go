package main

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.uber.org/zap"
)

func newTracerProvider(serviceName, zipkinEndpoint, otlpEndpoint string) (*sdktrace.TracerProvider, error) {
	zipkinExporter, err := zipkin.New(zipkinEndpoint)
	if err != nil {
		return nil, fmt.Errorf("create zipkin exporter: %w", err)
	}

	otlpExporter, err := otlptracehttp.New(
		context.Background(),
		otlptracehttp.WithEndpoint(otlpEndpoint),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("create otlp exporter: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(zipkinExporter),
		sdktrace.WithBatcher(otlpExporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(serviceName),
		)),
	)

	return tp, nil
}

func NewTracer(log *zap.Logger, serviceName, zipkinEndpoint, otlpEndpoint string) (*sdktrace.TracerProvider, error) {
	tp, err := newTracerProvider(serviceName, zipkinEndpoint, otlpEndpoint)
	if err != nil {
		return nil, err
	}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	otel.SetErrorHandler(otel.ErrorHandlerFunc(func(err error) {
		log.Error("otel error", zap.Error(err))
	}))
	otel.SetTracerProvider(tp)

	return tp, nil
}

func NewClientTracerProvider(serviceName, zipkinEndpoint, otlpEndpoint string) (*sdktrace.TracerProvider, error) {
	return newTracerProvider(serviceName, zipkinEndpoint, otlpEndpoint)
}
