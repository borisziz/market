package domain

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrCantReserveItem = errors.New("can not reserve item")
)

func (d *domain) CreateOrder(ctx context.Context, user int64, items []OrderItem) (int64, error) {
	order := &Order{Status: StatusNew, User: user, Items: items}
	orderID, err := d.OrdersRepository.CreateOrder(ctx, order)
	if err != nil {
		return 0, errors.Wrap(err, "create order")
	}
	order.ID = orderID
	go func() {
		err = d.TransactionManager.RunTransaction(context.Background(), func(ctxTX context.Context) error {
			var reserveFrom []ReservedItem
			for _, item := range order.Items {
				stocks, err := d.OrdersRepository.Stocks(ctxTX, item.Sku)
				if err != nil {
					return errors.Wrap(err, "check stocks")
				}
				var counter uint64 = 0
				for _, stock := range stocks {
					counter += stock.Count
					if counter > uint64(item.Count) {
						reserveFrom = append(reserveFrom, ReservedItem{
							WarehouseID: stock.WarehouseID,
							OrderItem: OrderItem{
								Sku:   item.Sku,
								Count: uint16(item.Count) - uint16(counter-stock.Count),
							}})
					} else {
						reserveFrom = append(reserveFrom, ReservedItem{
							WarehouseID: stock.WarehouseID,
							OrderItem: OrderItem{
								Sku:   item.Sku,
								Count: uint16(stock.Count),
							}})
					}
					if counter == uint64(item.Count) {
						break
					}
				}
				if counter < uint64(item.Count) {
					return ErrCantReserveItem
				}
			}
			for _, v := range reserveFrom {
				err = d.OrdersRepository.ReserveStock(ctxTX, order.ID, v)
				if err != nil {
					return errors.Wrap(err, "reserve stock")
				}
			}
			err = d.OrdersRepository.UpdateOrderStatus(ctxTX, order.ID, StatusAwaitingPayment, StatusNew)
			if err != nil {
				return errors.Wrap(err, "set order status")
			}
			return nil
		})
		if err != nil {
			err = d.OrdersRepository.UpdateOrderStatus(context.Background(), order.ID, StatusFailed, StatusNew)
			//TODO: do something with errors
		}
	}()
	return order.ID, nil
}
