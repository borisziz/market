package domain

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrNoCart        = errors.New("user has no cart")
	ErrNoSku         = errors.New("no sku in cart")
	ErrNoSoManyItems = errors.New("no so many items")
)

func (m *domain) DeleteFromCart(ctx context.Context, user int64, sku uint32, count uint16) error {
	return nil
}
