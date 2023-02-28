package domain

import "context"

type Stock struct {
	WarehouseID int64
	Count       uint64
}

func (d *domain) Stocks(ctx context.Context, sku uint32) ([]Stock, error) {
	return []Stock{
		{
			WarehouseID: 123,
			Count:       5,
		}}, nil
}
