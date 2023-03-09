package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"route256/libs/interceptors"
	transactor "route256/libs/postgres_transactor"
	"route256/loms/internal/api/loms/v1"
	"route256/loms/internal/config"
	"route256/loms/internal/domain"
	repository "route256/loms/internal/repository/postgres"
	desc "route256/loms/pkg/loms/v1"
	"sync"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
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
	tm, err := transactor.New(config.ConfigData.DBConnectURL)
	if err != nil {
		log.Fatal("init transaction manager: ", err)
	}
	repo := repository.NewItemsRepo(tm)
	desc.RegisterLOMSV1Server(grpcServer, loms.New(domain.New(repo, tm)))
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
	err := desc.RegisterLOMSV1HandlerFromEndpoint(ctx, mux, config.ConfigData.Ports.Grpc, opts)
	if err != nil {
		return err
	}

	log.Printf("http server running on port %v\n", config.ConfigData.Ports.Http)

	return http.ListenAndServe(config.ConfigData.Ports.Http, mux)
}
