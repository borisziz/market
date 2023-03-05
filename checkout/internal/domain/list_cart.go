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

func (m *domain) ListCart(ctx context.Context, user int64) ([]CartItem, error) {
	var items []CartItem
	for i, sku := range []uint32{1076963, 1148162, 1625903, 2618151, 2956315} {
		info, err := m.productServiceCaller.GetProduct(ctx, sku)
		if err != nil {
			return nil, errors.WithMessage(err, "getting product info")
		}
		items = append(items, CartItem{Sku: sku, Count: uint16(i % 3), ProductInfo: info})
	}
	return items, nil
}
