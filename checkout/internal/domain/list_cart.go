package domain

import (
	"context"

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
	for i, item := range items {
		info, err := d.productServiceCaller.GetProduct(ctx, item.Sku)
		if err != nil {
			return nil, errors.WithMessage(err, "getting product info")
		}
		items[i].ProductInfo = info
	}
	return items, nil
}
