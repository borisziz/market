package stocks

import (
	"context"
	"log"
	"route256/loms/internal/domain"

	"github.com/pkg/errors"
)

type Request struct {
	SKU uint32 `json:"sku"`
}

func (r Request) Validate() error {
	// TODO: implement
	return nil
}

type Item struct {
	WarehouseID int64  `json:"warehouseID"`
	Count       uint64 `json:"count"`
}

type Response struct {
	Stocks []Item `json:"stocks"`
}

type Handler struct {
	domain *domain.Domain
}

func New(domain *domain.Domain) *Handler {
	return &Handler{domain: domain}
}

func (h *Handler) Handle(ctx context.Context, request Request) (Response, error) {
	log.Printf("stocks: %+v", request)
	var response Response
	stocks, err := h.domain.Stocks(ctx, request.SKU)
	if err != nil {
		return response, errors.Wrap(err, "get stocks")
	}
	for _, v := range stocks {
		response.Stocks = append(response.Stocks, Item{v.WarehouseID, v.Count})
	}
	return response, nil
}
