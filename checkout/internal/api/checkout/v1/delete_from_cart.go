package checkout

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	desc "route256/checkout/pkg/checkout/v1"
)

func (i *Implementation) DeleteFromCart(ctx context.Context, req *desc.DeleteFromCartRequest) (*emptypb.Empty, error) {
	err := i.checkoutService.DeleteFromCart(ctx, req.User, req.Sku, uint16(req.Count))
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
