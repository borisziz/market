package loms

import (
	"context"
	"route256/checkout/internal/domain"
	loms "route256/loms/pkg/loms/v1"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type Client struct {
	c loms.LOMSV1Client
}

func New(conn *grpc.ClientConn) *Client {
	c := loms.NewLOMSV1Client(conn)
	return &Client{
		c: c,
	}
}

func (c *Client) Stocks(ctx context.Context, sku uint32) ([]domain.Stock, error) {
	request := &loms.StocksRequest{Sku: sku}
	response, err := c.c.Stocks(ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "client request")
	}
	stocks := make([]domain.Stock, 0, len(response.Stocks))
	for _, stock := range response.Stocks {
		stocks = append(stocks, domain.Stock{
			WarehouseID: stock.GetWarehouseID(),
			Count:       stock.GetCount(),
		})
	}
	return stocks, nil
}

func (c *Client) CreateOrder(ctx context.Context, user int64, cartItems []domain.CartItem) (int64, error) {
	request := &loms.CreateOrderRequest{User: user}
	for _, v := range cartItems {
		request.Items = append(request.Items, &loms.Item{Sku: v.Sku, Count: uint32(v.Count)})
	}
	response, err := c.c.CreateOrder(ctx, request)
	if err != nil {
		return 0, errors.Wrap(err, "client request")
	}
	return response.GetOrderID(), nil
}
