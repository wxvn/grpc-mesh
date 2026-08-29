package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wxvn/grpc-mesh/internal/common/config"
	"github.com/wxvn/grpc-mesh/internal/common/interceptors"
	"github.com/wxvn/grpc-mesh/internal/common/logger"
	"github.com/wxvn/grpc-mesh/internal/url/client"
	"github.com/wxvn/grpc-mesh/internal/url/service"
	"github.com/wxvn/grpc-mesh/internal/url/storage"
	urlgrpc "github.com/wxvn/grpc-mesh/internal/url/transport/grpc"
	urlhttp "github.com/wxvn/grpc-mesh/internal/url/transport/http"
	urlpb "github.com/wxvn/grpc-mesh/proto/url"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.MustLoad()
	log := logger.Setup(cfg.Env)
	ctx := context.Background()

	authConn, err := grpc.NewClient(
		cfg.AuthGRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Error("create auth grpc connection", "error", err)
		os.Exit(1)
	}
	defer authConn.Close()

	authClient := client.NewAuthClient(authConn)

	urlStorage, err := storage.New(ctx, cfg.URLPostgresURL)
	if err != nil {
		log.Error("create url storage", "error", err)
		os.Exit(1)
	}
	defer urlStorage.Close()

	urlService := service.NewURLService(urlStorage)

	middleware := urlhttp.NewMiddleware(authClient)
	httpServer := urlhttp.NewServer(urlService, middleware)

	grpcServer := urlgrpc.NewServer(
		urlService,
		cfg.URLPublicScheme,
		cfg.URLPublicHost,
		cfg.URLPublicPort,
	)

	grpcListener, err := net.Listen("tcp", ":"+cfg.URLGRPCPort)
	if err != nil {
		log.Error("listen grpc", "error", err)
		os.Exit(1)
	}
	defer grpcListener.Close()

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.Recovery(log),
			interceptors.RequestID(),
			interceptors.Logging(log, "url"),
			interceptors.Auth(authClient),
		),
	)

	urlpb.RegisterURLShortenerServiceServer(grpcSrv, grpcServer)
	reflection.Register(grpcSrv)

	httpListener, err := net.Listen("tcp", ":"+cfg.URLHTTPPort)
	if err != nil {
		log.Error("listen http", "error", err)
		os.Exit(1)
	}
	defer httpListener.Close()

	httpSrv := &http.Server{
		Handler: httpServer.Handler(),
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh

		log.Info("shutting down url servers")

		grpcDone := make(chan struct{})

		go func() {
			grpcSrv.GracefulStop()
			close(grpcDone)
		}()

		select {
		case <-grpcDone:
			log.Info("grpc server stopped")
		case <-time.After(5 * time.Second):
			log.Warn("grpc graceful shutdown timeout")
			grpcSrv.Stop()
		}

		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if err := httpSrv.Shutdown(shutdownCtx); err != nil {
			log.Error("http graceful shutdown", "error", err)
		}

		log.Info("http server stopped")
	}()

	go func() {
		log.Info(
			"http server started",
			"port",
			cfg.URLHTTPPort,
			"env",
			cfg.Env,
		)

		if err := httpSrv.Serve(httpListener); err != nil &&
			err != http.ErrServerClosed {
			log.Error("http serve", "error", err)
			os.Exit(1)
		}
	}()

	log.Info(
		"grpc server started",
		"port",
		cfg.URLGRPCPort,
		"env",
		cfg.Env,
	)

	if err := grpcSrv.Serve(grpcListener); err != nil {
		log.Error("grpc serve", "error", err)
		os.Exit(1)
	}
}
