package loms

import (
	"context"
	desc "route256/loms/pkg/loms/v1"

	"google.golang.org/protobuf/types/known/emptypb"
)

func (i *Implementation) OrderPayed(ctx context.Context, req *desc.OrderPayedRequest) (*emptypb.Empty, error) {
	err := i.lOMSService.OrderPayed(ctx, req.GetOrderID())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
