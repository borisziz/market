package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"route256/checkout/internal/api/checkout/v1"
	"route256/checkout/internal/clients/loms"
	"route256/checkout/internal/clients/productservice"
	"route256/checkout/internal/config"
	"route256/checkout/internal/domain"
	repository "route256/checkout/internal/repository/postgres"
	desc "route256/checkout/pkg/checkout/v1"
	"route256/libs/interceptors"
	transactor "route256/libs/postgres_transactor"
	"sync"
	"time"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
			log.Fatal("failed run grpc: ", err)
		}
	}()
	go func() {
		defer wg.Done()

		err := runHTTP(ctx)
		if err != nil {
			log.Fatal("failed run http: ", err)
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
	connProducts, err := grpc.Dial(config.ConfigData.Services.Products, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal("failed create products client: failed to connect to server:", err)
	}
	defer connProducts.Close()
	tm, err := transactor.New(config.ConfigData.DBConnectURL)
	if err != nil {
		log.Fatal("init transaction manager: ", err)
	}
	repo := repository.NewCartsRepo(tm)

	lomsClient := loms.New(connLoms)
	limiter := rate.NewLimiter(rate.Every(time.Second/10), 10)
	productsServiceClient := productservice.New(config.ConfigData.Token, connProducts)
	businessLogic, err := domain.New(lomsClient, productsServiceClient, repo, tm, limiter)
	if err != nil {
		log.Fatal("init business logic", err)
	}

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
		return errors.Wrap(err, "register handler")
	}

	log.Printf("http server running on port %v\n", config.ConfigData.Ports.Http)

	return http.ListenAndServe(config.ConfigData.Ports.Http, mux)
}
