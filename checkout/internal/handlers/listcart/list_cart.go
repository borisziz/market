package listcart

import (
	"context"
	"errors"
	"route256/checkout/internal/domain"
)

type Handler struct {
	businessLogic *domain.Domain
}

func New(businessLogic *domain.Domain) *Handler {
	return &Handler{
		businessLogic: businessLogic,
	}
}

type Request struct {
	User int64 `json:"user"`
}

var (
	ErrEmptyUser = errors.New("empty user")
)

func (r Request) Validate() error {
	if r.User == 0 {
		return ErrEmptyUser
	}
	return nil
}

type Item struct {
	Sku   uint32 `json:"sku"`
	Count uint16 `json:"count"`
	Name  string `json:"name"`
	Price uint32 `json:"price"`
}

type Response struct {
	Items      []Item `json:"items"`
	TotalPrice uint32 `json:"totalPrice"`
}

func (h *Handler) Handle(ctx context.Context, req Request) (Response, error) {
	var response Response

	cart, err := h.businessLogic.ListCart(ctx, req.User)
	if err != nil {
		return response, err
	}
	for _, item := range cart {
		response.Items = append(response.Items, Item{Sku: item.Sku, Count: item.Count, Name: item.Name, Price: item.Price})
		response.TotalPrice += item.Price * uint32(item.Count)
	}
	return response, nil
}
