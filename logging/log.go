package logging

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxLoggerKey struct{}

func newLogger() *zap.Logger {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, err := loggerConfig.Build()
	if err != nil {
		fmt.Printf("failed to initialize logger: %v", err)
		backupLogger, _ := zap.NewProduction()
		return backupLogger
	}
	return logger
}

func NewContextWithLogger(id string) context.Context {
	logger := newLogger().With(zap.String("loggerId", id))
	return context.WithValue(context.Background(), ctxLoggerKey{}, logger)
}

func LoggerFromContext(ctx context.Context) *zap.Logger {
	logger, ok := ctx.Value(ctxLoggerKey{}).(*zap.Logger)
	if !ok {
		fmt.Printf("failed to get logger from context")
		return newLogger()
	}
	return logger
}

func SugaredLoggerFromContext(ctx context.Context) *zap.SugaredLogger {
	return LoggerFromContext(ctx).Sugar()
}
