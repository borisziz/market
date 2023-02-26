package listorder

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
	Status string `json:"status"`
	User   int64  `json:"user"`
	Items  []Item `json:"items"`
}

type Item struct {
	Sku   uint32 `json:"sku"`
	Count uint16 `json:"count"`
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
	order, err := h.domain.ListOrder(ctx, request.OrderID)
	if err != nil {
		return response, errors.Wrap(err, "list order")
	}
	response.User = order.User
	response.Status = order.Status
	for _, v := range order.Items {
		response.Items = append(response.Items, Item{v.Sku, v.Count})
	}
	return response, nil
}
