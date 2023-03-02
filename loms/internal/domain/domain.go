package domain

import (
	"context"
	"errors"
	"github.com/brianvoe/gofakeit/v6"
)

var _ Domain = (*domain)(nil)

type ProductServiceCaller interface {
	GetProduct(ctx context.Context, sku uint32) (string, error)
}

type Domain interface {
	CreateOrder(ctx context.Context, user int64, items []OrderItem) (int64, error)
	ListOrder(ctx context.Context, orderID int64) (*Order, error)
	CancelOrder(ctx context.Context, orderID int64) error
	Stocks(ctx context.Context, sku uint32) ([]Stock, error)
	OrderPayed(ctx context.Context, orderID int64) error
}

type domain struct {
}

type SKUs map[int64]struct{}

func New() *domain {
	return &domain{}
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

const (
	StatusNew             = "new"
	StatusAwaitingPayment = "awaiting payment"
	StatusFailed          = "failed"
	StatusPayed           = "Payed"
	StatusCancelled       = "Cancelled"
	StatusUndefined       = "Undefined"
)

func createOrder(order Order) int64 {
	return 5
}

func setOrderStatus(orderID int64, status string) error {
	return nil
}

var ErrOrderNotFound = errors.New("order not found")

func getOrder(orderID int64) (*Order, error) {
	if false {
		return nil, ErrOrderNotFound
	}
	defaultItems := []OrderItem{{Sku: 1076963, Count: gofakeit.Uint16()}, {Sku: 1148162, Count: gofakeit.Uint16()}}
	defaultOrder := &Order{ID: orderID, Status: StatusAwaitingPayment, User: gofakeit.Int64(), Items: defaultItems}
	return defaultOrder, nil
}
