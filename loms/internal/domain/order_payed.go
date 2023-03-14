package domain

import (
	"context"

	"github.com/pkg/errors"
)

func (d *domain) OrderPayed(ctx context.Context, orderID int64) error {
	err := d.TransactionManager.RunTransaction(ctx, isoLevelSerializable, func(ctxTX context.Context) error {
		err := d.OrdersRepository.UpdateOrderStatus(ctxTX, orderID, StatusPayed, StatusAwaitingPayment)
		if err != nil {
			return errors.Wrap(err, "update status")
		}
		err = d.OrdersRepository.RemoveSoldedItems(ctxTX, orderID)
		if err != nil {
			return errors.Wrap(err, "remove solded items")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "order payed")
	}
	return nil
}
