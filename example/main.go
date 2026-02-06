package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"
)

const (
	ServerServiceName = "example-server"
	ClientServiceName = "example-client"
	ServerPort        = "8000"
	ZipkinEndpoint    = "http://localhost:9411/api/v2/spans"
	OTLPEndpoint      = "localhost:4318"
	GraylogAddr       = "localhost:12201"
	TickerInterval    = 15 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// logger
	logger := NewLogger(ServerServiceName, GraylogAddr)
	defer logger.Sync()

	// server tracing
	serverTP, err := NewTracer(logger, ServerServiceName, ZipkinEndpoint, OTLPEndpoint)
	if err != nil {
		logger.Fatal("failed to initialize server tracer", zap.Error(err))
	}
	defer func() {
		if err = serverTP.Shutdown(context.Background()); err != nil {
			logger.Error("failed to shutdown server tracer", zap.Error(err))
		}
	}()

	// client tracing
	clientTP, err := NewClientTracerProvider(ClientServiceName, ZipkinEndpoint, OTLPEndpoint)
	if err != nil {
		logger.Fatal("failed to initialize client tracer", zap.Error(err))
	}
	defer func() {
		if err = clientTP.Shutdown(context.Background()); err != nil {
			logger.Error("failed to shutdown client tracer", zap.Error(err))
		}
	}()

	// run server
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/data", MakeDataHandler(logger))
	mux.HandleFunc("GET /api/slow", MakeSlowHandler(logger))
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := MetricsMiddleware(otelhttp.NewHandler(mux, "example-server"))

	server := &http.Server{
		Addr:         ":" + ServerPort,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		logger.Info("starting server", zap.String("addr", server.Addr))
		if sErr := server.ListenAndServe(); sErr != nil && !errors.Is(sErr, http.ErrServerClosed) {
			logger.Fatal("server failed", zap.Error(sErr))
		}
	}()

	// run client
	httpClient := NewHTTPClient(clientTP)
	go RunLoadGenerator(ctx, logger, httpClient, clientTP, ServerPort, TickerInterval)

	<-ctx.Done()
	logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err = server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", zap.Error(err))
	}
	logger.Info("server stopped")
}
