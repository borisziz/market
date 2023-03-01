package loms

import (
	"context"
	"route256/loms/internal/domain"
	desc "route256/loms/pkg/loms/v1"
)

func (i *Implementation) ListOrder(ctx context.Context, req *desc.ListOrderRequest) (*desc.ListOrderResponse, error) {
	order, err := i.lOMSService.ListOrder(ctx, req.GetOrderID())
	if err != nil {
		return nil, err
	}
	items := make([]*desc.Item, 0, len(order.Items))
	for _, item := range order.Items {
		items = append(items, &desc.Item{
			Sku:   item.Sku,
			Count: uint32(item.Count),
		})
	}

	return &desc.ListOrderResponse{
		Status: StatusToStatusCode(order.Status),
		User:   order.User,
		Items:  items,
	}, nil
}

func StatusToStatusCode(status string) desc.OrderStatus {
	switch status {
	case domain.StatusNew:
		return desc.OrderStatus_New
	case domain.StatusAwaitingPayment:
		return desc.OrderStatus_AwaitingPayment
	case domain.StatusFailed:
		return desc.OrderStatus_Failed
	case domain.StatusPayed:
		return desc.OrderStatus_Payed
	case domain.StatusCancelled:
		return desc.OrderStatus_Cancelled
	default:
		return desc.OrderStatus_Undefined
	}
}
