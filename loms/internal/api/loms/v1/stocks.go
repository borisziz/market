package loms

import (
	"context"
	desc "route256/loms/pkg/loms/v1"
)

func (i *Implementation) Stocks(ctx context.Context, req *desc.StocksRequest) (*desc.StocksResponse, error) {
	stocks, err := i.lOMSService.Stocks(ctx, req.GetSku())
	if err != nil {
		return nil, err
	}
	res := make([]*desc.Stock, 0, len(stocks))
	for _, stock := range stocks {
		res = append(res, &desc.Stock{
			WarehouseID: stock.WarehouseID,
			Count:       stock.Count,
		})
	}

	return &desc.StocksResponse{Stocks: res}, nil
}
