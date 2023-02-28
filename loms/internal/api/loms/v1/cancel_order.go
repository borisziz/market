package loms

import (
	"context"
	"google.golang.org/protobuf/types/known/emptypb"
	desc "route256/loms/pkg/loms/v1"
)

func (i *Implementation) CancelOrder(ctx context.Context, req *desc.CancelOrderRequest) (*emptypb.Empty, error) {
	err := i.lOMSService.CancelOrder(ctx, req.GetUser())
	if err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
