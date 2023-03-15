package loms

import (
	"context"
	desc "route256/loms/pkg/loms/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) CancelOrder(ctx context.Context, req *desc.CancelOrderRequest) (*emptypb.Empty, error) {
	err := i.lOMSService.CancelOrder(ctx, req.GetOrderID())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
