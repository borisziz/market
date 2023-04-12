package domain

import (
	"context"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"
	"route256/libs/logger"
	desc "route256/loms/pkg/loms/v1"
)

type domain struct{}

func New() *domain {
	return &domain{}
}

func (d *domain) ReceiveOrder(data []byte) {
	var order desc.Order
	err := protojson.Unmarshal(data, &order)
	if err != nil {
		logger.Error(context.Background(), "Unmarshal order", zap.Error(err))
		return
	}
	logger.Info("receive order", zap.Any("order", order.String()))
}
