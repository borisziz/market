package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"route256/libs/interceptors"
	"route256/libs/kafka"
	transactor "route256/libs/postgres_transactor"
	"route256/loms/internal/api/loms/v1"
	"route256/loms/internal/config"
	"route256/loms/internal/domain"
	repository "route256/loms/internal/repository/postgres"
	"route256/loms/internal/sender"
	desc "route256/loms/pkg/loms/v1"
	"sync"
	"syscall"
	"time"

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
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		err := runGRPC(ctx)
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

func runGRPC(ctx context.Context) error {
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
		log.Fatal("init transaction manager:", err)
	}
	repo := repository.NewItemsRepo(tm)
	producer, err := kafka.NewSyncProducer(config.ConfigData.Kafka.Brokers)
	if err != nil {
		log.Fatal("init kafka producer:", err)
	}
	ns := sender.NewOrderSender(producer, config.ConfigData.Kafka.Topic)
	desc.RegisterLOMSV1Server(grpcServer, loms.New(domain.New(repo, tm, ns)))
	log.Printf("grps server running on port %v\n", config.ConfigData.Ports.Grpc)

	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			log.Fatal("failed to serve:", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down grpc server")
	grpcServer.GracefulStop()
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
	httpServer := &http.Server{
		Handler: mux,
		Addr:    config.ConfigData.Ports.Http,
	}
	go func() {
		err = httpServer.ListenAndServe()
		if err != nil {
			log.Fatal("failed to serve:", err)
		}
	}()
	<-ctx.Done()
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	log.Println("shutting down http server")
	return httpServer.Shutdown(ctxShutdown)
}
