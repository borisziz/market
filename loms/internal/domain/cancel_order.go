package domain

import (
	"context"

	"github.com/pkg/errors"
)

func (d *domain) CancelOrder(ctx context.Context, orderID int64) error {
	err := d.TransactionManager.RunTransaction(ctx, isoLevelSerializable, func(ctxTX context.Context) error {
		err := d.OrdersRepository.UpdateOrderStatus(ctxTX, orderID, StatusCancelled, StatusAwaitingPayment)
		if err != nil {
			return errors.Wrap(err, "set order status")
		}
		err = d.OrdersRepository.UnReserveItems(ctxTX, orderID)
		if err != nil {
			return errors.Wrap(err, "set items sold")
		}
		return nil
	})
	if err != nil {
		return errors.Wrap(err, "cancel order")
	}
	return nil
}
