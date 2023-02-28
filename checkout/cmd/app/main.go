package main

import (
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
	"log"
	"net"
	"route256/checkout/internal/api/checkout/v1"
	"route256/checkout/internal/clients/loms"
	"route256/checkout/internal/clients/productservice"
	"route256/checkout/internal/config"
	"route256/checkout/internal/domain"
	desc "route256/checkout/pkg/checkout/v1"
)

func main() {
	err := config.Init()
	if err != nil {
		log.Fatal("config init", err)
	}

	lis, err := net.Listen("tcp", config.ConfigData.Port)
	if err != nil {
		log.Fatalf("failed listen tcp at %v port", config.ConfigData.Port)
	}

	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(grpcValidator.UnaryServerInterceptor()))

	//clients
	lomsClient := loms.New(config.ConfigData.Services.Loms)
	productsServiceClient := productservice.New(config.ConfigData.Token, config.ConfigData.Services.Products)
	businessLogic := domain.New(lomsClient, productsServiceClient)

	desc.RegisterCheckoutV1Server(grpcServer, checkout.New(businessLogic))

	log.Printf("grps server running on port %v\n", config.ConfigData.Port)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
