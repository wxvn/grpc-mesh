package service

import (
	"context"
	"errors"

	"github.com/wxvn/grpc-mesh/internal/weather/domain"
)

type Location struct {
	City        string
	Coordinates *domain.Coordinates
}

func (s *WeatherService) GetCurrentWeather(
	ctx context.Context,
	location Location,
) (*domain.CurrentWeather, error) {
	var coordinates domain.Coordinates
	var city string

	switch {
	case location.City != "":
		var err error
		city = location.City
		coordinates, err = s.geocoding.GetCoordinates(ctx, location.City)
		if err != nil {
			return nil, err
		}
	case location.Coordinates != nil:
		coordinates = *location.Coordinates
	default:
		return nil, errors.New("location is required")
	}

	weather, err := s.weather.GetCurrentWeather(ctx, coordinates)
	if err != nil {
		return nil, err
	}

	elevation, err := s.elevation.GetElevation(ctx, coordinates)
	if err != nil {
		return nil, err
	}

	weather.City = city
	weather.Elevation = elevation

	return weather, nil
}
