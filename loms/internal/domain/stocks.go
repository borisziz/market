package domain

import (
	"context"

	"github.com/pkg/errors"
)

type Stock struct {
	WarehouseID int64
	Count       uint64
}

func (d *domain) Stocks(ctx context.Context, sku uint32) ([]Stock, error) {
	stocks, err := d.OrdersRepository.Stocks(ctx, sku)
	if err != nil {
		return nil, errors.Wrap(err, "get stocks")
	}
	return stocks, nil
}
