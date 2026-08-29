package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	authgrpc "github.com/wxvn/grpc-mesh/internal/auth/grpc"
	"github.com/wxvn/grpc-mesh/internal/auth/service"
	"github.com/wxvn/grpc-mesh/internal/auth/storage"

	"github.com/wxvn/grpc-mesh/internal/common/config"
	"github.com/wxvn/grpc-mesh/internal/common/interceptors"
	"github.com/wxvn/grpc-mesh/internal/common/logger"
	"github.com/wxvn/grpc-mesh/internal/common/password"
	"github.com/wxvn/grpc-mesh/internal/common/token"

	pb "github.com/wxvn/grpc-mesh/proto/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.MustLoad()
	log := logger.Setup(cfg.Env)

	ctx := context.Background()

	store, err := storage.New(
		ctx,
		cfg.AuthPostgresURL,
	)
	if err != nil {
		log.Error("create storage", "error", err)
		os.Exit(1)
	}
	defer store.Close()

	passwordHasher := password.NewHasher(5)

	tokenManager := token.NewManager(
		cfg.JWTSecret,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
		cfg.JWTIssuer,
	)

	authService := service.NewAuthService(
		store,
		store,
		passwordHasher,
		tokenManager,
	)

	authServer := authgrpc.NewServer(authService)

	lis, err := net.Listen(
		"tcp",
		":"+cfg.AuthGRPCPort,
	)
	if err != nil {
		log.Error("listen", "error", err)
		os.Exit(1)
	}

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RequestID(),
			interceptors.Logging(log, "auth"),
			interceptors.Recovery(log),
		),
	)

	pb.RegisterAuthServiceServer(
		grpcSrv,
		authServer,
	)

	reflection.Register(grpcSrv)

	sigCh := make(chan os.Signal, 1)

	signal.Notify(
		sigCh,
		os.Interrupt,
		syscall.SIGTERM,
	)

	go func() {
		<-sigCh

		log.Info("shutting down gRPC server")

		grpcSrv.GracefulStop()

		log.Info("gRPC server stopped")
	}()

	log.Info(
		"gRPC server started",
		"port", cfg.AuthGRPCPort,
		"env", cfg.Env,
	)

	if err := grpcSrv.Serve(lis); err != nil {
		log.Error("serve", "error", err)
		os.Exit(1)
	}
}
