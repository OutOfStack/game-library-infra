package main

import (
	"log/slog"
	"os"

	gelf "github.com/snovichkov/zap-gelf"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(serviceName, graylogAddr string) *zap.Logger {
	hostname, _ := os.Hostname()

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.RFC3339TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder

	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.Lock(os.Stderr),
		zapcore.InfoLevel,
	)

	cores := []zapcore.Core{consoleCore}

	graylogCore, err := gelf.NewCore(
		gelf.Addr(graylogAddr),
		gelf.Host(hostname),
	)
	if err != nil {
		slog.Error("failed to create graylog core", "error", err)
	} else {
		cores = append(cores, graylogCore)
	}

	logger := zap.New(zapcore.NewTee(cores...)).With(
		zap.String("service", serviceName),
	)

	return logger
}
