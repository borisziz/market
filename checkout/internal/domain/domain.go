package domain

import (
	"context"

	"github.com/pkg/errors"
)

var _ Domain = (*domain)(nil)

const (
	isoLevelSerializable    = "serializable"
	isoLevelRepeatableRead  = "repeatable read"
	isoLevelReadCommitted   = "read committed"
	isoLevelReadUncommitted = "read uncommitted"
)

type TransactionManager interface {
	RunTransaction(ctx context.Context, isoLevel string, f func(ctxTX context.Context) error) error
}

type CartsRepository interface {
	GetCartItem(ctx context.Context, user int64, sku uint32) (*CartItem, error)
	AddToCart(ctx context.Context, user int64, sku uint32, count uint16) error
	DeleteFromCart(ctx context.Context, user int64, sku uint32, count uint16, full bool) error
	GetCart(ctx context.Context, user int64) ([]CartItem, error)
	DeleteCart(ctx context.Context, user int64) error
}

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
	GetSKUs(ctx context.Context) (SKUs, error)
}

type Limiter interface {
	Wait(ctx context.Context) error
}

type domain struct {
	lOMSCaller           LOMSCaller
	productServiceCaller ProductServiceCaller
	rateLimiter          Limiter
	repo                 CartsRepository
	tm                   TransactionManager
	skus                 SKUs
}

type SKUs map[uint32]struct{}

func New(lOMSCaller LOMSCaller, productServiceCaller ProductServiceCaller, repo CartsRepository, tm TransactionManager, limiter Limiter) (*domain, error) {
	skus, err := productServiceCaller.GetSKUs(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "init skus")
	}
	return &domain{
		lOMSCaller:           lOMSCaller,
		productServiceCaller: productServiceCaller,
		rateLimiter:          limiter,
		repo:                 repo,
		tm:                   tm,
		skus:                 skus,
	}, nil
}
