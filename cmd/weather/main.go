package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wxvn/grpc-mesh/internal/common/config"
	"github.com/wxvn/grpc-mesh/internal/common/interceptors"
	"github.com/wxvn/grpc-mesh/internal/common/logger"

	"github.com/wxvn/grpc-mesh/internal/weather/client"
	weathergrpc "github.com/wxvn/grpc-mesh/internal/weather/grpc"
	"github.com/wxvn/grpc-mesh/internal/weather/service"

	weatherpb "github.com/wxvn/grpc-mesh/proto/weather"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	cfg := config.MustLoad()
	log := logger.Setup(cfg.Env)

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

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	weatherClient := client.NewWeatherClient(httpClient)
	elevationClient := client.NewElevationClient(httpClient)
	geocodingClient := client.NewGeocodingClient(httpClient)

	weatherService := service.NewWeatherService(
		weatherClient,
		elevationClient,
		geocodingClient,
	)

	weatherServer := weathergrpc.NewServer(weatherService)

	lis, err := net.Listen(
		"tcp",
		":"+cfg.WeatherGRPCPort,
	)
	if err != nil {
		log.Error("listen", "error", err)
		os.Exit(1)
	}
	defer lis.Close()

	grpcSrv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.Recovery(log),
			interceptors.RequestID(),
			interceptors.Logging(log, "weather"),
			interceptors.Auth(authClient),
		),
	)

	weatherpb.RegisterWeatherServiceServer(grpcSrv, weatherServer)
	reflection.Register(grpcSrv)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh

		log.Info("shutting down gRPC server")

		done := make(chan struct{})

		go func() {
			grpcSrv.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			log.Info("gRPC server stopped")
		case <-time.After(5 * time.Second):
			log.Warn("graceful shutdown timeout")
			grpcSrv.Stop()
		}
	}()

	log.Info(
		"gRPC server started",
		"port", cfg.WeatherGRPCPort,
		"env", cfg.Env,
	)

	if err := grpcSrv.Serve(lis); err != nil {
		log.Error("serve", "error", err)
		os.Exit(1)
	}
}
