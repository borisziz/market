package domain

import (
	"context"
)

var _ Domain = (*domain)(nil)

type ProductServiceCaller interface {
	GetProduct(ctx context.Context, sku uint32) (string, error)
}

const (
	isoLevelSerializable    = "serializable"
	isoLevelRepeatableRead  = "repeatable read"
	isoLevelReadCommitted   = "read committed"
	isoLevelReadUncommitted = "read uncommitted"
)

type TransactionManager interface {
	RunTransaction(ctx context.Context, isoLevel string, f func(ctxTX context.Context) error) error
}

type OrdersRepository interface {
	GetOrder(ctx context.Context, id int64) (*Order, error)
	CreateOrder(ctx context.Context, order *Order) (int64, error)
	UpdateOrderStatus(ctx context.Context, id int64, status string, statusBefore string) error
	ReserveStock(ctx context.Context, orderID int64, item ReservedItem) error
	UnReserveItems(ctx context.Context, orderID int64) error
	RemoveSoldedItems(ctx context.Context, orderID int64) error
	Stocks(ctx context.Context, sku uint32) ([]Stock, error)
}

type Deps struct {
	OrdersRepository
	TransactionManager
}

type Domain interface {
	CreateOrder(ctx context.Context, user int64, items []OrderItem) (int64, error)
	ListOrder(ctx context.Context, orderID int64) (*Order, error)
	CancelOrder(ctx context.Context, orderID int64) error
	Stocks(ctx context.Context, sku uint32) ([]Stock, error)
	OrderPayed(ctx context.Context, orderID int64) error
}

type domain struct {
	Deps
}

type SKUs map[int64]struct{}

func New(repo OrdersRepository, tm TransactionManager) *domain {
	return &domain{Deps{repo, tm}}
}

type OrderItem struct {
	Sku   uint32
	Count uint16
}

type Order struct {
	ID     int64
	Status string
	User   int64
	Items  []OrderItem
}

type ReservedItem struct {
	OrderItem
	WarehouseID int64
}

const (
	StatusNew             = "new"
	StatusAwaitingPayment = "awaiting payment"
	StatusFailed          = "failed"
	StatusPayed           = "payed"
	StatusCancelled       = "cancelled"
	StatusUndefined       = "undefined"
)
