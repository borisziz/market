package domain

import (
	"context"
	"errors"
)

type ProductServiceCaller interface {
	GetProduct(ctx context.Context, sku uint32) (string, error)
}

type Domain struct {
}

type SKUs map[int64]struct{}

func New() *Domain {
	return &Domain{}
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
	defaultItems := []OrderItem{{Sku: 1076963, Count: 1}, {Sku: 1148162, Count: 3}}
	defaultOrder := &Order{ID: orderID, Status: StatusAwaitingPayment, User: 3, Items: defaultItems}
	return defaultOrder, nil
}
