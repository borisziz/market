package domain

import (
	"context"

	"github.com/pkg/errors"
)

func (d *domain) OrderPayed(ctx context.Context, orderID int64) error {
	err := d.TransactionManager.RunTransaction(ctx, isoLevelSerializable, func(ctxTX context.Context) error {
		order, err := d.OrdersRepository.GetOrder(ctxTX, orderID)
		if err != nil {
			return errors.Wrap(err, "get order")
		}
		if order.Status != StatusAwaitingPayment {
			return errWrongStatus
		}
		err = d.OrdersRepository.UpdateOrderStatus(ctxTX, orderID, StatusPayed, order.Status)
		if err != nil {
			return errors.Wrap(err, "update status")
		}
		err = d.OrdersRepository.RemoveSoldItems(ctxTX, orderID)
		if err != nil {
			return errors.Wrap(err, "remove sold items")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "order payed")
	}
	return nil
}
