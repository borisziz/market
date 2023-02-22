package productservice

import (
	"context"
	"route256/checkout/internal/domain"
	"route256/libs/clientwrapper"

	"github.com/pkg/errors"
)

type Client struct {
	token string
	url   string

	urlProducts string
	urlSKUs     string
}

func New(token, url string) *Client {
	return &Client{
		token:       token,
		url:         url,
		urlProducts: url + "/get_product",
		urlSKUs:     url + "/list_skus",
	}
}

type ProductsRequest struct {
	Token string `json:"token"`
	SKU   uint32 `json:"sku"`
}

type ProductsResponse struct {
	Name  string `json:"name"`
	Price uint32 `json:"price"`
}

func (c *Client) GetProduct(ctx context.Context, sku uint32) (domain.ProductInfo, error) {
	request := ProductsRequest{Token: c.token, SKU: sku}
	var response ProductsResponse
	w := clientwrapper.New(c.urlProducts, request, response)
	var productInfo domain.ProductInfo
	response, err := w.ClientRequest(ctx)
	if err != nil {
		return productInfo, errors.Wrap(err, "client request")
	}
	productInfo.Name = response.Name
	productInfo.Price = response.Price

	return productInfo, nil
}

type SKUsRequest struct {
	Token         string `json:"token"`
	StartAfterSku int64  `json:"startAfterSku"`
	Count         int64  `json:"count"`
}

type SKUsResponse struct {
	SKUs []int64 `json:"skus"`
}

func (c *Client) GetSKUs(ctx context.Context) (domain.SKUs, error) {
	request := SKUsRequest{Token: c.token, StartAfterSku: 0, Count: 1000}
	var response SKUsResponse
	w := clientwrapper.New(c.urlProducts, request, response)
	response, err := w.ClientRequest(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "client request")
	}
	skus := make(domain.SKUs, len(response.SKUs))
	for _, sku := range response.SKUs {
		skus[sku] = struct{}{}
	}
	return skus, nil
}
