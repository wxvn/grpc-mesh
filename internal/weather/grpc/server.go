package grpc

import (
	"context"

	"github.com/wxvn/grpc-mesh/internal/weather/domain"
	"github.com/wxvn/grpc-mesh/internal/weather/service"
	pb "github.com/wxvn/grpc-mesh/proto/weather"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WeatherService interface {
	GetCurrentWeather(ctx context.Context, location service.Location) (*domain.CurrentWeather, error)
	GetForecast(ctx context.Context, location service.Location, days int) (*domain.Forecast, error)
}

type WeatherServer struct {
	pb.UnimplementedWeatherServiceServer
	service WeatherService
}

func NewServer(service WeatherService) *WeatherServer {
	return &WeatherServer{service: service}
}

func (s *WeatherServer) GetCurrentWeather(ctx context.Context, req *pb.GetCurrentWeatherRequest) (*pb.GetCurrentWeatherResponse, error) {
	location, err := parseCurrentLocation(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.service.GetCurrentWeather(ctx, location)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.GetCurrentWeatherResponse{
		City:        result.City,
		Latitude:    result.Coordinates.Latitude,
		Longitude:   result.Coordinates.Longitude,
		Elevation:   result.Elevation,
		Temperature: result.Temperature,
		FeelsLike:   result.FeelsLike,
		Humidity:    result.Humidity,
		WindSpeed:   result.WindSpeed,
		Condition:   string(result.Condition),
		Time:        result.Time,
		Timezone:    result.Timezone,
	}, nil
}

func (s *WeatherServer) GetForecast(ctx context.Context, req *pb.GetForecastRequest) (*pb.GetForecastResponse, error) {
	location, err := parseForecastLocation(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	result, err := s.service.GetForecast(ctx, location, int(req.Days))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	days := make([]*pb.ForecastDay, 0, len(result.Days))
	for _, day := range result.Days {
		days = append(days, &pb.ForecastDay{
			Date:                     day.Date,
			TempMin:                  day.TempMin,
			TempMax:                  day.TempMax,
			PrecipitationProbability: day.PrecipitationProbability,
			Condition:                string(day.Condition),
		})
	}

	return &pb.GetForecastResponse{
		City:      result.City,
		Latitude:  result.Coordinates.Latitude,
		Longitude: result.Coordinates.Longitude,
		Elevation: result.Elevation,
		Days:      days,
	}, nil
}

func parseCurrentLocation(req *pb.GetCurrentWeatherRequest) (service.Location, error) {
	if req.GetCity() != "" {
		return service.Location{City: req.GetCity()}, nil
	}

	if coordinates := req.GetCoordinates(); coordinates != nil {
		return service.Location{
			Coordinates: &domain.Coordinates{
				Latitude:  coordinates.Latitude,
				Longitude: coordinates.Longitude,
			},
		}, nil
	}

	return service.Location{}, status.Error(codes.InvalidArgument, "location is required")
}

func parseForecastLocation(req *pb.GetForecastRequest) (service.Location, error) {
	if req.GetCity() != "" {
		return service.Location{City: req.GetCity()}, nil
	}

	if coordinates := req.GetCoordinates(); coordinates != nil {
		return service.Location{
			Coordinates: &domain.Coordinates{
				Latitude:  coordinates.Latitude,
				Longitude: coordinates.Longitude,
			},
		}, nil
	}

	return service.Location{}, status.Error(codes.InvalidArgument, "location is required")
}
