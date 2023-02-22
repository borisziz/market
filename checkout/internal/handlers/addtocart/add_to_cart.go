package addtocart

import (
	"context"
	"errors"
	"log"
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
	User  int64  `json:"user"`
	Sku   uint32 `json:"sku"`
	Count uint16 `json:"count"`
}

var (
	ErrEmptyUser = errors.New("empty user")
	ErrEmptySKU  = errors.New("empty sku")
	ErrZeroCount = errors.New("count must be greater than 0")
)

func (r Request) Validate() error {
	if r.User == 0 {
		return ErrEmptyUser
	}
	if r.Sku == 0 {
		return ErrEmptySKU
	}
	if r.Count == 0 {
		return ErrZeroCount
	}
	return nil
}

type Response struct{}

func (h *Handler) Handle(ctx context.Context, req Request) (Response, error) {
	log.Printf("addToCart: %+v", req)

	var response Response

	err := h.businessLogic.AddToCart(ctx, req.User, req.Sku, req.Count)
	if err != nil {
		return response, err
	}

	return response, nil
}
