package service

import (
	"context"
	"errors"

	"github.com/wxvn/grpc-mesh/internal/weather/domain"
)

func (s *WeatherService) GetForecast(ctx context.Context, location Location, days int) (*domain.Forecast, error) {
	if days <= 0 {
		return nil, errors.New("days must be greater than zero")
	}

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

	daysData, err := s.weather.GetForecast(ctx, coordinates, days)
	if err != nil {
		return nil, err
	}

	elevation, err := s.elevation.GetElevation(ctx, coordinates)
	if err != nil {
		return nil, err
	}

	return &domain.Forecast{
		City:        city,
		Coordinates: coordinates,
		Elevation:   elevation,
		Days:        daysData,
	}, nil
}
