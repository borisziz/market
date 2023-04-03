package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net/http"
	"os"
	"os/signal"
	"route256/checkout/internal/api/checkout/v1"
	"route256/checkout/internal/clients/loms"
	"route256/checkout/internal/clients/productservice"
	"route256/checkout/internal/config"
	"route256/checkout/internal/domain"
	repository "route256/checkout/internal/repository/postgres"
	desc "route256/checkout/pkg/checkout/v1"
	"route256/libs/interceptors"
	"route256/libs/limiter"
	"route256/libs/logger"
	transactor "route256/libs/postgres_transactor"
	"sync"
	"syscall"
	"time"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcValidator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger.Init(true)
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
			logger.Fatal("failed run grpc: ", zap.Error(err))
		}
	}()
	go func() {
		defer wg.Done()

		err := runHTTP(ctx)
		if err != nil {
			logger.Fatal("failed run http: ", zap.Error(err))
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

	//clients
	connLoms, err := grpc.Dial(config.ConfigData.Services.Loms, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("failed create loms client: failed to connect to server:", zap.Error(err))
	}
	defer connLoms.Close()
	connProducts, err := grpc.Dial(config.ConfigData.Services.Products, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal("failed create products client: failed to connect to server:", zap.Error(err))
	}
	defer connProducts.Close()
	tm, err := transactor.New(config.ConfigData.DBConnectURL)
	if err != nil {
		logger.Fatal("init transaction manager: ", zap.Error(err))
	}
	repo := repository.NewCartsRepo(tm)

	lomsClient := loms.New(connLoms)
	//limiter := rate.NewLimiter(rate.Every(time.Second/10), 15)
	limiter := limiter.NewLimiter(10, 15)
	productsServiceClient := productservice.New(config.ConfigData.Token, connProducts)
	poolConfig := domain.PoolConfig{
		AmountWorkers:     config.ConfigData.WorkerPool.Workers,
		MaxRetries:        config.ConfigData.WorkerPool.Retries,
		WithCancelOnError: config.ConfigData.WorkerPool.WithCancelOnError,
	}
	businessLogic, err := domain.New(lomsClient, productsServiceClient, repo, tm, limiter, poolConfig)
	if err != nil {
		logger.Fatal("init business logic", zap.Error(err))
	}

	desc.RegisterCheckoutV1Server(grpcServer, checkout.New(businessLogic))

	logger.Info("grps server running on port %v\n", zap.String("addr", config.ConfigData.Ports.Grpc))

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

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	err := desc.RegisterCheckoutV1HandlerFromEndpoint(ctx, mux, config.ConfigData.Ports.Grpc, opts)
	if err != nil {
		return errors.Wrap(err, "register handler")
	}

	logger.Info("http server running on port %v\n", zap.String("addr", config.ConfigData.Ports.Http))
	httpServer := &http.Server{
		Handler: mux,
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
