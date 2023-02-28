package loms

import (
	"route256/loms/internal/domain"
	desc "route256/loms/pkg/loms/v1"
)

type Implementation struct {
	desc.UnimplementedLOMSV1Server

	lOMSService domain.Domain
}

func New(lOMSService domain.Domain) *Implementation {
	return &Implementation{
		desc.UnimplementedLOMSV1Server{},
		lOMSService,
	}
}
