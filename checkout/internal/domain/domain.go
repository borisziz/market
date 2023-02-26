package domain

import "context"

type LOMSCaller interface {
	Stocks(ctx context.Context, sku uint32) ([]Stock, error)
	CreateOrder(ctx context.Context, user int64, cartItems []CartItem) (OrderID, error)
}

type ProductServiceCaller interface {
	GetProduct(ctx context.Context, sku uint32) (ProductInfo, error)
}

type Domain struct {
	lOMSCaller           LOMSCaller
	productServiceCaller ProductServiceCaller
}

type SKUs map[int64]struct{}

func New(lOMSCaller LOMSCaller, productServiceCaller ProductServiceCaller) *Domain {
	return &Domain{
		lOMSCaller:           lOMSCaller,
		productServiceCaller: productServiceCaller,
	}
}
