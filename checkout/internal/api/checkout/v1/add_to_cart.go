package checkout

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	desc "route256/checkout/pkg/checkout/v1"
)

func (i *Implementation) AddToCart(ctx context.Context, req *desc.AddToCartRequest) (*emptypb.Empty, error) {
	err := i.checkoutService.AddToCart(ctx, req.GetUser(), req.GetSku(), uint16(req.GetCount()))
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
