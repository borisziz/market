package domain

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrNoCart        = errors.New("user has no cart")
	ErrNoSku         = errors.New("no sku in cart")
	ErrNoSoManyItems = errors.New("no so many items in cart")
)

func (d *domain) DeleteFromCart(ctx context.Context, user int64, sku uint32, count uint16) error {
	err := d.tm.RunTransaction(ctx, func(ctxTX context.Context) error {
		item, err := d.repo.GetCartItem(ctxTX, user, sku)
		if err != nil && !errors.Is(err, ErrNoSameItemsInCart) {
			return errors.Wrap(err, "get cart item")
		}
		if count > item.Count {
			return ErrNoSoManyItems
		}
		err = d.repo.DeleteFromCart(ctxTX, user, sku, count, count == item.Count)
		if err != nil {
			return errors.Wrap(err, "delete from cart")
		}
		return nil
	})
	return err
}
