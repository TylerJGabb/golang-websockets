package logging

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewStdoutLogger returns a new logger that logs to stdout.
func NewStdoutLogger() logr.Logger {
	return funcr.New(func(prefix string, args string) {
		if prefix != "" {
			fmt.Printf("%s: %s\n", prefix, args)
		} else {
			fmt.Printf("%s\n", args)
		}
	}, funcr.Options{})
}

func demoLogr() {
	l := NewStdoutLogger()
	l.Info("default info log", "stringVal", "value", "intVal", 12345)
	l.V(0).Info("v(0) info log", "stringVal", "value", "intVal", 12345)
	l.Error(fmt.Errorf("error"), "error log", "stringVal", "value", "intVal", 12345)
	l.Info("default info log", "stringVal", "value", "intVal", 12345)
}

func demoZapSugar() {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	sugar.Info("failed to fetch URL",
		// Structured context as strongly typed Field values.
		zap.String("url", "some--url"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)

	sugar.Infow("failed to fetch URL",
		// Structured context as a variadic list of Field values.
		"url", "some--url",
		"attempt", 3,
		"backoff", time.Second,
	)
}

func demoZapCustomTimstampFormat() {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = "timestamp"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder
	logger, err := loggerConfig.Build()
	if err != nil {
		fmt.Printf("failed to initialize logger: %v", err)
		return
	}
	defer logger.Sync()
	sugar := logger.With(zap.String("app", "demo")).Sugar()

	logger.Info("failed to fetch URL",
		zap.Time("time", time.Now()),
	)
	sugar.Infow("failed to fetch URL",
		// Structured context as a variadic list of Field values.
		"url", "some--url",
		"attempt", 3,
		"backoff", time.Minute,
		"nil", nil,
	)
}
