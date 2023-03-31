package domain

import (
	"context"

	"github.com/pkg/errors"
)

var errWrongStatus = errors.New("wrong status")

func (d *domain) CancelOrder(ctx context.Context, orderID int64) error {
	var order *Order
	var err error
	err = d.TransactionManager.RunTransaction(ctx, isoLevelSerializable, func(ctxTX context.Context) error {
		order, err = d.OrdersRepository.GetOrder(ctxTX, orderID)
		if err != nil {
			return errors.Wrap(err, "get order")
		}
		if order.Status != StatusAwaitingPayment {
			return errWrongStatus
		}
		err = d.OrdersRepository.UpdateOrderStatus(ctxTX, orderID, StatusCancelled, order.Status)
		if err != nil {
			return errors.Wrap(err, "set order status")
		}
		order.Status = StatusCancelled
		err = d.OrdersRepository.UnReserveItems(ctxTX, orderID)
		if err != nil {
			return errors.Wrap(err, "set items sold")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "cancel order")
	}
	err = d.NotificationsSender.SendOrder(order)
	if err != nil {
		return errors.Wrap(err, "send order")
	}
	return nil
}
