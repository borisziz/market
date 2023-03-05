package loms

import (
	"context"
	"route256/loms/internal/domain"
	desc "route256/loms/pkg/loms/v1"
)

func (i *Implementation) CreateOrder(ctx context.Context, req *desc.CreateOrderRequest) (*desc.CreateOrderResponse, error) {
	items := make([]domain.OrderItem, 0, len(req.GetItems()))
	for _, item := range req.GetItems() {
		items = append(items, domain.OrderItem{
			Sku:   item.GetSku(),
			Count: uint16(item.GetCount()),
		})
	}
	orderID, err := i.lOMSService.CreateOrder(ctx, req.GetUser(), items)
	if err != nil {
		return nil, err
	}

	return &desc.CreateOrderResponse{OrderID: orderID}, nil
}
