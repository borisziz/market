package domain

import (
	"context"

	"github.com/pkg/errors"
)

var ErrNotItemsInCart = errors.New("no items in cart")

func (d *domain) Purchase(ctx context.Context, user int64) (int64, error) {
	items, err := d.repo.GetCart(ctx, user)
	if err != nil {
		return 0, errors.Wrap(err, "get cart")
	}
	if len(items) == 0 {
		return 0, ErrNotItemsInCart
	}
	orderID, err := d.lOMSCaller.CreateOrder(ctx, user, items)
	if err != nil {
		return 0, errors.WithMessage(err, "creating order")
	}
	err = d.repo.DeleteCart(ctx, user)
	if err != nil {
		return orderID, errors.Wrap(err, "delete cart after create order")
	}
	return orderID, nil
}
