package interceptors

import (
	"context"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var HistogramClientResponseTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "homework",
	Subsystem: "grpc",
	Name:      "histogram_client_response_time_seconds",
	Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 16),
},
	[]string{"service", "success"},
)

func ClientInterceptor(serviceName string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, resp interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption) error {

		var parentCtx opentracing.SpanContext
		if parent := opentracing.SpanFromContext(ctx); parent != nil {
			parentCtx = parent.Context()
		}
		span := opentracing.GlobalTracer().StartSpan(
			method,
			opentracing.ChildOf(parentCtx),
			ext.SpanKindRPCClient,
		)
		defer span.Finish()
		span.SetTag(string(ext.Component), "gRPC")
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}
		mdWriter := mdReaderWriter{md}
		err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, mdWriter)
		if err != nil {
			span.LogFields(log.String("event", "Tracer.Inject() failed"), log.Error(err))
		}
		ctx = metadata.NewOutgoingContext(ctx, md)

		timeStart := time.Now()
		err = invoker(ctx, method, req, resp, cc, opts...)
		if err != nil {
			HistogramClientResponseTime.WithLabelValues(serviceName, operationStatusFailed).Observe(time.Since(timeStart).Seconds())
			ext.Error.Set(span, true)
		} else {
			HistogramClientResponseTime.WithLabelValues(serviceName, operationStatusSuccess).Observe(time.Since(timeStart).Seconds())
		}
		return err
	}
}
