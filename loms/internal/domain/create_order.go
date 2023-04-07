package domain

import (
	"context"
	"route256/libs/logger"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	ErrCantReserveItem = errors.New("can not reserve item")
)

func (d *domain) CreateOrder(ctx context.Context, user int64, items []OrderItem) (int64, error) {
	order := &Order{Status: StatusNew, User: user, Items: items}
	err := d.TransactionManager.RunTransaction(ctx, isoLevelRepeatableRead, func(ctxTX context.Context) error {
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
	err = d.NotificationsSender.SendOrder(order)
	if err != nil {
		return 0, errors.Wrap(err, "send order")
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = d.TransactionManager.RunTransaction(ctx, isoLevelRepeatableRead, func(ctxTX context.Context) error {
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
								Count: item.Count - uint16(counter-stock.Count),
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
			order.Status = StatusAwaitingPayment
			err = d.NotificationsSender.SendOrder(order)
			if err != nil {
				logger.Error(ctxTX, "send order", zap.Error(err))
			}
			return nil
		})
		if err != nil {
			logger.Error(ctx, "error create order", zap.Error(err))
			err = d.OrdersRepository.UpdateOrderStatus(ctx, order.ID, StatusFailed, order.Status)
			if err != nil {
				logger.Error(ctx, "error update order status", zap.Error(err))
				return
			}
			order.Status = StatusFailed
			err = d.NotificationsSender.SendOrder(order)
			if err != nil {
				logger.Error(ctx, "send order", zap.Error(err))
			}
		}
	}()
	time.AfterFunc(10*time.Minute, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := d.OrdersRepository.UpdateOrderStatus(ctx, order.ID, StatusCancelled, StatusAwaitingPayment)
		if err != nil && !errors.Is(err, ErrOrderNotFound) {
			logger.Error(ctx, "error update order status", zap.Error(err))
			return
		}
		order.Status = StatusCancelled
		err = d.NotificationsSender.SendOrder(order)
		if err != nil {
			logger.Error(ctx, "send order", zap.Error(err))
		}
	})
	return order.ID, nil
}
