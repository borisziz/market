package domain

import (
	"context"
	"github.com/brianvoe/gofakeit/v6"
)

type Stock struct {
	WarehouseID int64
	Count       uint64
}

func (d *domain) Stocks(ctx context.Context, sku uint32) ([]Stock, error) {
	return []Stock{
		{
			WarehouseID: gofakeit.Int64(),
			Count:       gofakeit.Uint64(),
		}}, nil
}
