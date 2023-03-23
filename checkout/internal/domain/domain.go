package domain

//go:generate sh -c "rm ./zzz*"
//go:generate minimock -i CartsRepository -o "./zzz_carts_repo_minimock_test.go"
//go:generate minimock -i LOMSCaller -o "./zzz_loms_minimock_test.go"
//go:generate minimock -i ProductServiceCaller -o "./zzz_products_minimock_test.go"
//go:generate minimock -i TransactionManager -o "./zzz_tm_minimock_test.go"
//go:generate minimock -i Limiter -o "./zzz_limiter_minimock_test.go"

import (
	"context"
	"time"

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
	poolConfig           PoolConfig
	skus                 SKUs
}

type PoolConfig struct {
	AmountWorkers     uint16
	MaxRetries        uint8
	WithCancelOnError bool
}

type SKUs map[uint32]struct{}

func New(lOMSCaller LOMSCaller, productServiceCaller ProductServiceCaller, repo CartsRepository, tm TransactionManager, limiter Limiter, poolConfig PoolConfig) (*domain, error) {
	d := &domain{
		lOMSCaller:           lOMSCaller,
		productServiceCaller: productServiceCaller,
		rateLimiter:          limiter,
		repo:                 repo,
		tm:                   tm,
		poolConfig:           poolConfig,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	skus, err := d.initSkus(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "init skus")
	}
	d.skus = skus
	return d, nil
}

func (d *domain) initSkus(ctx context.Context) (SKUs, error) {
	skus, err := d.productServiceCaller.GetSKUs(ctx)
	return skus, err
}

func NewMock(deps ...interface{}) (*domain, error) {
	d := &domain{}

	for _, v := range deps {
		switch s := v.(type) {
		case CartsRepository:
			d.repo = s
		case LOMSCaller:
			d.lOMSCaller = s
		case ProductServiceCaller:
			d.productServiceCaller = s
			skus, err := d.initSkus(context.Background())
			if err != nil {
				return nil, errors.Wrap(err, "init skus")
			}
			d.skus = skus
		case Limiter:
			d.rateLimiter = s
		case TransactionManager:
			d.tm = s
		case PoolConfig:
			d.poolConfig = s
		}
	}
	return d, nil
}
