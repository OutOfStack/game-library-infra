package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func MakeDataHandler(log *zap.Logger) http.HandlerFunc {
	tracer := otel.Tracer("example-server")

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "processData",
			trace.WithAttributes(attribute.String("handler", "data")),
		)
		defer span.End()

		delay := time.Duration(10+rand.IntN(90)) * time.Millisecond
		simulateWork(ctx, tracer, delay)

		traceID := span.SpanContext().TraceID().String()
		log.Info("handled data request",
			zap.String("trace_id", traceID),
			zap.Duration("delay", delay),
			zap.String("path", r.URL.Path),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","trace_id":"%s","delay_ms":%d}`, traceID, delay.Milliseconds())
	}
}

func MakeSlowHandler(log *zap.Logger) http.HandlerFunc {
	tracer := otel.Tracer("example-server")

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, span := tracer.Start(r.Context(), "processSlowRequest",
			trace.WithAttributes(attribute.String("handler", "slow")),
		)
		defer span.End()

		delay := time.Duration(200+rand.IntN(800)) * time.Millisecond
		simulateWork(ctx, tracer, delay)

		if rand.IntN(10) == 0 {
			span.SetAttributes(attribute.Bool("error", true))
			log.Warn("slow request failed",
				zap.String("trace_id", span.SpanContext().TraceID().String()),
				zap.String("path", r.URL.Path),
			)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"random failure"}`))
			return
		}

		traceID := span.SpanContext().TraceID().String()
		log.Info("handled slow request",
			zap.String("trace_id", traceID),
			zap.Duration("delay", delay),
			zap.String("path", r.URL.Path),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ok","trace_id":"%s","delay_ms":%d}`, traceID, delay.Milliseconds())
	}
}

func simulateWork(ctx context.Context, tracer trace.Tracer, delay time.Duration) {
	_, span := tracer.Start(ctx, "simulateWork",
		trace.WithAttributes(attribute.Int64("delay_ms", delay.Milliseconds())),
	)
	defer span.End()

	select {
	case <-ctx.Done():
		return
	default:
		time.Sleep(delay)
	}
}
