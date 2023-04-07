package logger

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"go.uber.org/zap"
)

var globalLogger *zap.Logger

func Init(devel bool) {
	globalLogger = New(devel)
}

func New(devel bool) *zap.Logger {
	var logger *zap.Logger
	var err error
	if devel {
		logger, err = zap.NewDevelopment()
	} else {
		cfg := zap.NewProductionConfig()
		cfg.DisableCaller = true
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		logger, err = cfg.Build()
	}
	if err != nil {
		panic(err)
	}
	return logger
}

func Debug(msg string, fields ...zap.Field) {
	globalLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	globalLogger.Info(msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...zap.Field) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		if spanContext, ok := span.Context().(jaeger.SpanContext); ok {
			fields = append(fields,
				zap.String("trace", spanContext.TraceID().String()),
				zap.String("span", spanContext.SpanID().String()),
			)
		}
	}
	globalLogger.Error(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	globalLogger.Warn(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	globalLogger.Fatal(msg, fields...)
}
