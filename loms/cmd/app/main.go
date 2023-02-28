package main

import (
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"google.golang.org/grpc"
	"log"
	"net"
	"route256/loms/internal/api/loms/v1"
	"route256/loms/internal/config"
	"route256/loms/internal/domain"
	desc "route256/loms/pkg/loms/v1"
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

	desc.RegisterLOMSV1Server(grpcServer, loms.New(domain.New()))
	log.Printf("grps server running on port %v\n", config.ConfigData.Port)

	err = grpcServer.Serve(lis)
	if err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}
