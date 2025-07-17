package main

import (
	"batch-saver/api"
	"batch-saver/internal/config"
	"batch-saver/internal/grpc"
	"batch-saver/internal/service"
	"batch-saver/internal/storage"
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	ggrpc "google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	setupGracefulShutdown(cancel)

	cfg := config.Read()
	zerolog.SetGlobalLevel(cfg.GetLogLevel())

	if err := run(ctx, cfg); err != nil {
		log.Err(err).Msg("Got error")
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg config.Config) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCServerPort))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to listen")
	}
	grpcServer := ggrpc.NewServer()

	go func() {
		<-ctx.Done()
		grpcServer.GracefulStop()
	}()

	repo, err := storage.NewRepository(cfg.PostgresCfg)
	if err != nil {
		return err
	}

	wg := new(sync.WaitGroup)
	svc := service.NewService(ctx, wg, service.NewWriterPool(wg, repo, cfg.MaxConcurrentWrites), cfg.ServiceCfg)
	api.RegisterBatchSaverServiceServer(grpcServer, grpc.NewResolver(svc))
	grpcServer.Serve(lis)

	wg.Wait()
	return nil
}

func setupGracefulShutdown(stop func()) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signalChannel
		log.Info().Msg("Got interrupt signal")
		stop()
	}()
}
