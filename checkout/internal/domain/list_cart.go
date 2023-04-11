package domain

import (
	"context"
	"fmt"
	"route256/libs/pool"
	"time"

	"github.com/pkg/errors"
)

type ProductInfo struct {
	Name  string `json:"name"`
	Price uint32 `json:"price"`
}

type CartItem struct {
	Sku   uint32
	Count uint16
	ProductInfo
}

func (d *domain) ListCart(ctx context.Context, user int64) ([]CartItem, error) {

	items, err := d.repo.GetCart(ctx, user)
	if err != nil {
		return nil, errors.Wrap(err, "get cart")
	}
	wp, errorsChan := pool.NewPool(ctx, d.poolConfig.AmountWorkers, d.poolConfig.MaxRetries, d.poolConfig.WithCancelOnError)
	for i, item := range items {
		i := i
		item := item
		pi, ok := d.cache.Get(fmt.Sprintf("%d", item.Sku))
		if ok {
			items[i].ProductInfo = pi.(ProductInfo)
			continue
		}
		var task pool.Task
		task.Task = func() error {
			_ = d.rateLimiter.Wait(ctx)
			info, err := d.productServiceCaller.GetProduct(ctx, item.Sku)
			if err != nil {
				return err
			}
			//time.Sleep(time.Duration(i) * time.Second)
			items[i].ProductInfo = info
			d.cache.Set(fmt.Sprintf("%d", item.Sku), info, 10*time.Second)
			return nil
		}
		wp.Submit(task)
	}
	//После того как отправили все таски, не блокируя основную рутину, чтобы провалиться дальше в чтение канала ошибок, закрываем пул.
	go wp.Close()
	//Сделал так, чтобы при получении ошибки сразу выходить.
	//Выйдем из цикла, когда отработает закрытие пула после выполнения всех задач или по cancel
	for err := range errorsChan {
		return nil, errors.Wrap(err, "getting product info")
	}
	return items, nil
}
