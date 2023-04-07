package tracing

import (
	"route256/libs/logger"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"go.uber.org/zap"
)

func Init(serviceName string) {
	cfg, err := config.FromEnv()
	if err != nil {
		logger.Fatal("Cannot init tracing", zap.Error(err))
	}
	cfg.Sampler.Param = 1
	cfg.Sampler.Type = jaeger.SamplerTypeConst
	cfg.ServiceName = serviceName
	tracer, _, err := cfg.NewTracer()
	if err != nil {
		logger.Fatal("Cannot init tracing", zap.Error(err))
	}
	opentracing.SetGlobalTracer(tracer)
}
