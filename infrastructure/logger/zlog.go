package logger

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const timeEncoderOfLayout = "2006-01-02 15:04:05"

func Init() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(timeEncoderOfLayout)

	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to build logger: %w", err)
	}

	return logger, nil
}

type ctxLoggerKey struct{}

func WithContext(ctx context.Context, logger *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxLoggerKey{}, logger)
}

func FromContext(ctx context.Context) *zap.SugaredLogger {
	logger, ok := ctx.Value(ctxLoggerKey{}).(*zap.Logger)
	if !ok {
		defaultLogger, _ := zap.NewProduction()
		return defaultLogger.Sugar()
	}

	return logger.Sugar()
}
