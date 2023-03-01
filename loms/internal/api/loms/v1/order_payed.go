package loms

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	desc "route256/loms/pkg/loms/v1"
)

func (i *Implementation) OrderPayed(ctx context.Context, req *desc.OrderPayedRequest) (*emptypb.Empty, error) {
	err := i.lOMSService.OrderPayed(ctx, req.GetOrderID())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
