package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env string

	AuthPostgresURL string
	URLPostgresURL  string

	AuthGRPCPort    string
	WeatherGRPCPort string
	URLGRPCPort     string
	URLHTTPPort     string

	AuthGRPCAddress string

	URLPublicScheme string
	URLPublicHost   string
	URLPublicPort   string

	JWTSecret string
	JWTIssuer string

	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		fmt.Println(".env file not found, using environment variables")
	}

	return &Config{
		Env: getEnv("ENV", "local"),

		AuthPostgresURL: getEnv("AUTH_POSTGRES_URL", ""),
		URLPostgresURL:  getEnv("URL_POSTGRES_URL", ""),

		AuthGRPCPort:    getEnv("AUTH_GRPC_PORT", "50051"),
		WeatherGRPCPort: getEnv("WEATHER_GRPC_PORT", "50052"),
		URLGRPCPort:     getEnv("URL_GRPC_PORT", "50053"),
		URLHTTPPort:     getEnv("URL_HTTP_PORT", "8080"),

		AuthGRPCAddress: getEnv(
			"AUTH_GRPC_ADDRESS",
			"localhost:50051",
		),

		URLPublicScheme: getEnv(
			"URL_PUBLIC_SCHEME",
			"http",
		),

		URLPublicHost: getEnv(
			"URL_PUBLIC_HOST",
			"localhost",
		),

		URLPublicPort: getEnv(
			"URL_PUBLIC_PORT",
			"8080",
		),

		JWTSecret: getEnv("JWT_SECRET", ""),
		JWTIssuer: getEnv("JWT_ISSUER", "auth-service"),

		AccessTokenTTL: mustDuration(
			"ACCESS_TOKEN_TTL",
			"15m",
		),

		RefreshTokenTTL: mustDuration(
			"REFRESH_TOKEN_TTL",
			"24h",
		),
	}
}

func requiredEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("%s is required", key))
	}

	return value
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func mustDuration(key, fallback string) time.Duration {
	value := getEnv(key, fallback)

	duration, err := time.ParseDuration(value)
	if err != nil {
		panic(fmt.Sprintf("invalid %s: %v", key, err))
	}

	return duration
}
