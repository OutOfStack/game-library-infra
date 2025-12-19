package main

import (
	"context"
	"math/rand/v2"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func NewHTTPClient(tp *sdktrace.TracerProvider) *http.Client {
	return &http.Client{
		Transport: otelhttp.NewTransport(
			http.DefaultTransport,
			otelhttp.WithTracerProvider(tp),
		),
		Timeout: 30 * time.Second,
	}
}

func RunLoadGenerator(ctx context.Context, log *zap.Logger, client *http.Client, tp *sdktrace.TracerProvider, serverPort string, tickerInterval time.Duration) {
	tracer := tp.Tracer("example-client")
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	baseURL := "http://localhost:" + serverPort
	endpoints := []string{"/api/data", "/api/slow"}

	log.Info("starting load generator", zap.String("base_url", baseURL))

	for {
		select {
		case <-ctx.Done():
			log.Info("stopping load generator")
			return
		case <-ticker.C:
			endpoint := endpoints[rand.IntN(len(endpoints))]
			url := baseURL + endpoint

			reqCtx, span := tracer.Start(ctx, "clientRequest",
				trace.WithAttributes(
					attribute.String("endpoint", endpoint),
					attribute.String("url", url),
				),
			)

			req, err := http.NewRequestWithContext(reqCtx, http.MethodGet, url, nil)
			if err != nil {
				log.Error("failed to create request", zap.Error(err))
				span.End()
				continue
			}

			resp, err := client.Do(req)
			if err != nil {
				log.Error("request failed",
					zap.String("url", url),
					zap.Error(err),
				)
				span.End()
				continue
			}

			log.Info("client request completed",
				zap.String("trace_id", span.SpanContext().TraceID().String()),
				zap.String("url", url),
				zap.Int("status", resp.StatusCode),
			)

			resp.Body.Close()
			span.End()
		}
	}
}
