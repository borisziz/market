package interceptors

import (
	"context"
	"strings"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	RequestsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "homework",
		Subsystem: "grpc",
		Name:      "requests_total",
	})
	ResponseCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "homework",
		Subsystem: "grpc",
		Name:      "responses_total",
	},
		[]string{"success"},
	)
	HistogramResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "homework",
		Subsystem: "grpc",
		Name:      "histogram_response_time_seconds",
		Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 16),
	},
		[]string{"success"},
	)
)

const (
	operationStatusSuccess = "true"
	operationStatusFailed  = "false"
)

func ServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	RequestsCounter.Inc()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}
	var span opentracing.Span
	if spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, mdReaderWriter{md}); err == nil {
		span, ctx = opentracing.StartSpanFromContext(ctx, info.FullMethod, ext.RPCServerOption(spanContext))
	} else {
		span, ctx = opentracing.StartSpanFromContext(ctx, info.FullMethod)
	}
	defer span.Finish()
	span.SetTag(string(ext.Component), "gRPC")

	timeStart := time.Now()

	res, err := handler(ctx, req)
	if err != nil {
		HistogramResponseTime.WithLabelValues(operationStatusFailed).Observe(time.Since(timeStart).Seconds())
		ResponseCounter.WithLabelValues(operationStatusFailed).Inc()
		ext.Error.Set(span, true)
	} else {
		HistogramResponseTime.WithLabelValues(operationStatusSuccess).Observe(time.Since(timeStart).Seconds())
		ResponseCounter.WithLabelValues(operationStatusSuccess).Inc()
	}
	return res, err
}

type mdReaderWriter struct {
	metadata.MD
}

func (w mdReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	w.MD[key] = append(w.MD[key], val)
}

func (w mdReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vals := range w.MD {
		for _, v := range vals {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}
