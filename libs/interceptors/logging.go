package interceptors

import (
	"context"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"route256/libs/logger"
)

// LoggingInterceptor ...
func LoggingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logger.Debug("incoming request", zap.String("method", info.FullMethod))

	res, err := handler(ctx, req)
	if err != nil {
		logger.Error("handler error", zap.String("method", info.FullMethod), zap.Error(err))
		return nil, err
	}
	logger.Debug("request succeed", zap.String("method", info.FullMethod))
	return res, nil
}
