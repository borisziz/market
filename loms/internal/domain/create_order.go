package domain

import (
	"context"
	"log"
	"time"

	"github.com/pkg/errors"
)

var (
	ErrCantReserveItem = errors.New("can not reserve item")
)

func (d *domain) CreateOrder(ctx context.Context, user int64, items []OrderItem) (int64, error) {
	order := &Order{Status: StatusNew, User: user, Items: items}
	err := d.TransactionManager.RunTransaction(context.Background(), isoLevelSerializable, func(ctxTX context.Context) error {
		orderID, err := d.OrdersRepository.CreateOrder(ctxTX, order)
		if err != nil {
			return errors.Wrap(err, "create order")
		}
		order.ID = orderID
		return nil
	})
	if err != nil {
		return 0, err
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = d.TransactionManager.RunTransaction(ctx, isoLevelSerializable, func(ctxTX context.Context) error {
			//TODO: Use worker pool
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
			err = d.OrdersRepository.UpdateOrderStatus(ctxTX, order.ID, StatusAwaitingPayment, order.Status)
			if err != nil {
				return errors.Wrap(err, "set order status")
			}
			return nil
		})
		if err != nil {
			log.Println("error create order", err)
			err = d.OrdersRepository.UpdateOrderStatus(ctx, order.ID, StatusFailed, order.Status)
			if err != nil {
				log.Println("error update order status", err)
			}
		}
	}()
	time.AfterFunc(10*time.Minute, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = d.OrdersRepository.UpdateOrderStatus(ctx, order.ID, StatusCancelled, StatusAwaitingPayment)
		if err != nil && !errors.Is(err, ErrOrderNotFound) {
			log.Println("error update order status", err)
		}
	})
	return order.ID, nil
}
