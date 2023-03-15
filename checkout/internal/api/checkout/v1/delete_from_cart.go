package checkout

import (
	"context"
	desc "route256/checkout/pkg/checkout/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) DeleteFromCart(ctx context.Context, req *desc.DeleteFromCartRequest) (*emptypb.Empty, error) {
	err := i.checkoutService.DeleteFromCart(ctx, req.GetUser(), req.GetSku(), uint16(req.GetCount()))
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
