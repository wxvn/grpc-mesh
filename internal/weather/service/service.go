package service

import (
	"context"

	"github.com/wxvn/grpc-mesh/internal/weather/domain"
)

type WeatherClient interface {
	GetCurrentWeather(ctx context.Context, coordinates domain.Coordinates) (*domain.CurrentWeather, error)
	GetForecast(ctx context.Context, coordinates domain.Coordinates, days int) ([]domain.ForecastDay, error)
}

type ElevationClient interface {
	GetElevation(ctx context.Context, coordinates domain.Coordinates) (float64, error)
}

type GeocodingClient interface {
	GetCoordinates(ctx context.Context, city string) (domain.Coordinates, error)
}

type WeatherService struct {
	weather   WeatherClient
	elevation ElevationClient
	geocoding GeocodingClient
}

func NewWeatherService(
	weather WeatherClient,
	elevation ElevationClient,
	geocoding GeocodingClient,
) *WeatherService {
	return &WeatherService{
		weather:   weather,
		elevation: elevation,
		geocoding: geocoding,
	}
}
