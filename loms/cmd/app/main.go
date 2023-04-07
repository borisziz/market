package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"route256/libs/interceptors"
	"route256/libs/kafka"
	"route256/libs/logger"
	transactor "route256/libs/postgres_transactor"
	"route256/libs/tracing"
	"route256/loms/internal/api/loms/v1"
	"route256/loms/internal/config"
	"route256/loms/internal/domain"
	repository "route256/loms/internal/repository/postgres"
	"route256/loms/internal/sender"
	desc "route256/loms/pkg/loms/v1"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger.Init(true)
	tracing.Init("loms")
	err := config.Init()
	if err != nil {
		logger.Fatal("config init", zap.Error(err))
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		err := runGRPC(ctx)
		if err != nil {
			logger.Fatal("run grpc", zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()

		err := runHTTP(ctx)
		if err != nil {
			logger.Fatal("run http", zap.Error(err))
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
				interceptors.ServerInterceptor,
				interceptors.LoggingInterceptor,
				grpcValidator.UnaryServerInterceptor(),
			),
		),
	)
	tm, err := transactor.New(config.ConfigData.DBConnectURL)
	if err != nil {
		logger.Fatal("init transaction manager:", zap.Error(err))
	}
	repo := repository.NewItemsRepo(tm)
	producer, err := kafka.NewSyncProducer(config.ConfigData.Kafka.Brokers)
	if err != nil {
		logger.Fatal("init kafka producer:", zap.Error(err))
	}
	ns := sender.NewOrderSender(producer, config.ConfigData.Kafka.Topic)
	desc.RegisterLOMSV1Server(grpcServer, loms.New(domain.New(repo, tm, ns)))
	logger.Info("grps server running on port", zap.String("addr", config.ConfigData.Ports.Grpc))

	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			logger.Fatal("failed to serve:", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down grpc server")
	grpcServer.GracefulStop()
	return nil
}

func runHTTP(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	r := mux.NewRouter()
	mux := runtime.NewServeMux()
	r.PathPrefix("/loms").Handler(mux)
	r.Handle("/metrics", promhttp.Handler())

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := desc.RegisterLOMSV1HandlerFromEndpoint(ctx, mux, config.ConfigData.Ports.Grpc, opts)
	if err != nil {
		return err
	}

	logger.Info("http server running on port", zap.String("addr", config.ConfigData.Ports.Http))
	httpServer := &http.Server{
		Handler: r,
		Addr:    config.ConfigData.Ports.Http,
	}
	go func() {
		err = httpServer.ListenAndServe()
		if err != nil {
			logger.Fatal("failed to serve:", zap.Error(err))
		}
	}()
	<-ctx.Done()
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	logger.Info("shutting down http server")
	return httpServer.Shutdown(ctxShutdown)
}
