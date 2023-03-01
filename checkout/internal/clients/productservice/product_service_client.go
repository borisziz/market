package productservice

import (
	"context"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"route256/checkout/internal/domain"
	product "route256/checkout/pkg/products"
)

type Client struct {
	token string
	c     product.ProductServiceClient
}

func New(token string, conn *grpc.ClientConn) *Client {
	c := product.NewProductServiceClient(conn)
	return &Client{
		token: token,
		c:     c,
	}
}

func (c *Client) GetProduct(ctx context.Context, sku uint32) (domain.ProductInfo, error) {
	request := &product.GetProductRequest{
		Token: c.token,
		Sku:   sku,
	}
	var productInfo domain.ProductInfo
	response, err := c.c.GetProduct(ctx, request)
	if err != nil {
		return productInfo, errors.Wrap(err, "client request")
	}
	productInfo.Name = response.GetName()
	productInfo.Price = response.GetPrice()

	return productInfo, nil
}

func (c *Client) GetSKUs(ctx context.Context) (domain.SKUs, error) {
	request := &product.ListSkusRequest{
		Token:         c.token,
		StartAfterSku: 0,
		Count:         1000,
	}
	response, err := c.c.ListSkus(ctx, request)
	if err != nil {
		return nil, errors.Wrap(err, "client request")
	}
	skus := make(domain.SKUs, len(response.GetSkus()))
	for _, sku := range response.GetSkus() {
		skus[sku] = struct{}{}
	}
	return skus, nil
}
