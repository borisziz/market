package loms

import (
	"context"
	"route256/checkout/internal/domain"
	"route256/libs/clientwrapper"

	"github.com/pkg/errors"
)

type Client struct {
	url string

	urlStocks      string
	urlCreateOrder string
}

func New(url string) *Client {
	return &Client{
		url: url,

		urlStocks:      url + "/stocks",
		urlCreateOrder: url + "/createOrder",
	}
}

type StocksRequest struct {
	SKU uint32 `json:"sku"`
}

type StocksItem struct {
	WarehouseID int64  `json:"warehouseID"`
	Count       uint64 `json:"count"`
}

type StocksResponse struct {
	Stocks []StocksItem `json:"stocks"`
}

func (c *Client) Stocks(ctx context.Context, sku uint32) ([]domain.Stock, error) {
	request := StocksRequest{SKU: sku}
	response, err := clientwrapper.ClientRequest[StocksRequest, StocksResponse]()(ctx, c.urlStocks, request)
	if err != nil {
		return nil, errors.Wrap(err, "client request")
	}
	stocks := make([]domain.Stock, 0, len(response.Stocks))
	for _, stock := range response.Stocks {
		stocks = append(stocks, domain.Stock{
			WarehouseID: stock.WarehouseID,
			Count:       stock.Count,
		})
	}
	return stocks, nil
}

type CreateOrderRequest struct {
	User  int64  `json:"user"`
	Items []Item `json:"items"`
}

type Item struct {
	Sku   uint32 `json:"sku"`
	Count uint16 `json:"count"`
}

type CreateOrderResponse struct {
	OrderID int64 `json:"orderID"`
}

func (c *Client) CreateOrder(ctx context.Context, user int64, cartItems []domain.CartItem) (domain.OrderID, error) {
	request := CreateOrderRequest{User: user}
	for _, v := range cartItems {
		request.Items = append(request.Items, Item{Sku: v.Sku, Count: v.Count})
	}
	var orderID domain.OrderID
	response, err := clientwrapper.ClientRequest[CreateOrderRequest, CreateOrderResponse]()(ctx, c.urlCreateOrder, request)
	if err != nil {
		return orderID, errors.Wrap(err, "client request")
	}
	orderID = domain.OrderID(response.OrderID)

	return orderID, nil
}
