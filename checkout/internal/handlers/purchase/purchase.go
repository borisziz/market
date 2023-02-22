package purchase

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

type Response struct {
	OrderID int64 `json:"orderID"`
}

func (h *Handler) Handle(ctx context.Context, req Request) (Response, error) {
	var response Response

	orderID, err := h.businessLogic.Purchase(ctx, req.User)
	if err != nil {
		return response, err
	}
	response.OrderID = int64(orderID)

	return response, nil
}
