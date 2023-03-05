package main

import (
	"context"
	"fmt"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net"
	"net/http"
	"route256/checkout/internal/api/checkout/v1"
	"route256/checkout/internal/clients/loms"
	"route256/checkout/internal/clients/productservice"
	"route256/checkout/internal/config"
	"route256/checkout/internal/domain"
	desc "route256/checkout/pkg/checkout/v1"
	"route256/libs/interceptors"
	"sync"
)

func main() {
	err := config.Init()
	if err != nil {
		log.Fatal("config init", err)
	}

	ctx := context.Background()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		err := runGRPC()
		if err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		defer wg.Done()

		err := runHTTP(ctx)
		if err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
}

func runGRPC() error {
	lis, err := net.Listen("tcp", config.ConfigData.Ports.Grpc)
	if err != nil {
		return fmt.Errorf("failed listen tcp at %v port", config.ConfigData.Ports.Grpc)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpcMiddleware.ChainUnaryServer(
				interceptors.LoggingInterceptor,
				grpcValidator.UnaryServerInterceptor(),
			),
		),
	)

	//clients
	connLoms, err := grpc.Dial(config.ConfigData.Services.Loms, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed create loms client: failed to connect to server:", err)
	}
	defer connLoms.Close()
	lomsClient := loms.New(connLoms)
	connProducts, err := grpc.Dial(config.ConfigData.Services.Products, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed create products client: failed to connect to server:", err)
	}
	defer connProducts.Close()
	productsServiceClient := productservice.New(config.ConfigData.Token, connProducts)
	businessLogic := domain.New(lomsClient, productsServiceClient)

	desc.RegisterCheckoutV1Server(grpcServer, checkout.New(businessLogic))

	log.Printf("grps server running on port %v\n", config.ConfigData.Ports.Grpc)

	err = grpcServer.Serve(lis)
	if err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
	return nil
}

func runHTTP(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := desc.RegisterCheckoutV1HandlerFromEndpoint(ctx, mux, config.ConfigData.Ports.Grpc, opts)
	if err != nil {
		return err
	}

	log.Printf("http server running on port %v\n", config.ConfigData.Ports.Http)

	return http.ListenAndServe(config.ConfigData.Ports.Http, mux)
}
