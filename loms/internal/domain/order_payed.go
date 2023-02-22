package domain

import (
	"context"

	"github.com/pkg/errors"
)

func (m *Domain) OrderPayed(ctx context.Context, orderID int64) error {
	order, err := getOrder(orderID)
	if err != nil {
		return errors.Wrap(err, "get order")
	}
	err = setItemsSold(order.Items)
	if err != nil {
		return errors.Wrap(err, "set items sold")
	}
	err = setOrderStatus(orderID, StatusPayed)
	if err != nil {
		return errors.Wrap(err, "set order status")
	}
	return nil
}

func setItemsSold(items []OrderItem) error {
	return nil
}
