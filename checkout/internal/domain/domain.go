package domain

import (
	"context"
)

var _ Domain = (*domain)(nil)

type Domain interface {
	AddToCart(ctx context.Context, user int64, sku uint32, count uint16) error
	DeleteFromCart(ctx context.Context, user int64, sku uint32, count uint16) error
	ListCart(ctx context.Context, user int64) ([]CartItem, error)
	Purchase(ctx context.Context, user int64) (int64, error)
}

type LOMSCaller interface {
	Stocks(ctx context.Context, sku uint32) ([]Stock, error)
	CreateOrder(ctx context.Context, user int64, cartItems []CartItem) (int64, error)
}

type ProductServiceCaller interface {
	GetProduct(ctx context.Context, sku uint32) (ProductInfo, error)
}

type domain struct {
	lOMSCaller           LOMSCaller
	productServiceCaller ProductServiceCaller
}

type SKUs map[uint32]struct{}

func New(lOMSCaller LOMSCaller, productServiceCaller ProductServiceCaller) *domain {
	return &domain{
		lOMSCaller:           lOMSCaller,
		productServiceCaller: productServiceCaller,
	}
}
