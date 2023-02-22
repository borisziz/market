package cancelorder

import (
	"context"
	"route256/loms/internal/domain"

	"github.com/pkg/errors"
)

type Request struct {
	OrderID int64 `json:"orderID"`
}

var (
	ErrEmptyOrderID = errors.New("empty orderID")
)

func (r Request) Validate() error {
	if r.OrderID == 0 {
		return ErrEmptyOrderID
	}
	return nil
}

type Response struct {
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
	var response Response
	err := h.domain.CancelOrder(ctx, request.OrderID)
	if err != nil {
		return response, errors.Wrap(err, "order payed")
	}
	return response, nil
}
