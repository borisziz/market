package domain

import (
	"context"

	"github.com/pkg/errors"
)

type Stock struct {
	WarehouseID int64
	Count       uint64
}

var (
	ErrInsufficientStocks = errors.New("insufficient stocks")
	ErrNoSameItemsInCart  = errors.New("no same items in cart")
	ErrInvalidSKU         = errors.New("invalid sku")
)

func (d *domain) AddToCart(ctx context.Context, user int64, sku uint32, count uint16) error {
	_, ok := d.skus[sku]
	if !ok {
		return ErrInvalidSKU
	}
	err := d.tm.RunTransaction(ctx, isoLevelRepeatableRead, func(ctxTX context.Context) error {
		item, err := d.repo.GetCartItem(ctxTX, user, sku)
		if err != nil && !errors.Is(err, ErrNoSameItemsInCart) {
			return errors.Wrap(err, "get cart item")
		}
		if errors.Is(err, ErrNoSameItemsInCart) {
			item = &CartItem{}
		}
		stocks, err := d.lOMSCaller.Stocks(ctxTX, sku)
		if err != nil {
			return errors.WithMessage(err, "checking stocks")
		}
		counter := int64(count + item.Count)
		for _, stock := range stocks {
			counter -= int64(stock.Count)
			if counter <= 0 {
				break
			}
		}
		if counter > 0 {
			return ErrInsufficientStocks
		}
		err = d.repo.AddToCart(ctxTX, user, sku, count)
		if err != nil {
			return errors.Wrap(err, "add to cart")
		}
		return nil
	})
	return err
}
