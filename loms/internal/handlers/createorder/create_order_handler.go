package createorder

import (
	"context"
	"route256/loms/internal/domain"

	"github.com/pkg/errors"
)

type Request struct {
	User  int64  `json:"user"`
	Items []Item `json:"items"`
}

type Item struct {
	Sku   uint32 `json:"sku"`
	Count uint16 `json:"count"`
}

var (
	ErrEmptyUser = errors.New("empty user")
	ErrEmptySKU  = errors.New("one of the items has empty sku")
	ErrZeroCount = errors.New("count all of the items must be greater than 0")
)

func (r Request) Validate() error {
	if r.User == 0 {
		return ErrEmptyUser
	}
	for _, item := range r.Items {
		if item.Sku == 0 {
			return ErrEmptySKU
		}
		if item.Count == 0 {
			return ErrZeroCount
		}
	}
	return nil
}

type Response struct {
	OrderID int64 `json:"orderID"`
}

type Handler struct {
	domain *domain.Domain
}

func New(domain *domain.Domain) *Handler {
	return &Handler{
		domain: domain,
	}
}

func (h *Handler) Handle(ctx context.Context, request Request) (Response, error) {
	var items []domain.OrderItem
	for _, v := range request.Items {
		items = append(items, domain.OrderItem{Sku: v.Sku, Count: v.Count})
	}
	var response Response
	orderID, err := h.domain.CreateOrder(ctx, request.User, items)
	if err != nil {
		return response, errors.Wrap(err, "create order")
	}
	response.OrderID = int64(orderID)
	return response, nil
}
