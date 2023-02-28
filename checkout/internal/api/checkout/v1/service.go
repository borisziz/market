package checkout

import (
	"route256/checkout/internal/domain"
	desc "route256/checkout/pkg/checkout/v1"
)

type Implementation struct {
	desc.UnimplementedCheckoutV1Server

	checkoutService domain.Domain
}

func New(checkoutService domain.Domain) *Implementation {
	return &Implementation{
		desc.UnimplementedCheckoutV1Server{},
		checkoutService,
	}
}
