package checkout

import (
	"context"
	desc "route256/checkout/pkg/checkout/v1"
)

func (i *Implementation) Purchase(ctx context.Context, req *desc.PurchaseRequest) (*desc.PurchaseResponse, error) {
	orderID, err := i.checkoutService.Purchase(ctx, req.User)
	if err != nil {
		return nil, err
	}

	return &desc.PurchaseResponse{OrderID: orderID}, nil
}
