package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/wxvn/grpc-mesh/internal/weather/domain"
)

const weatherAPIURL = "https://api.open-meteo.com/v1/forecast"

type WeatherClient struct {
	client *http.Client
}

func NewWeatherClient(client *http.Client) *WeatherClient {
	return &WeatherClient{client: client}
}

type currentWeatherResponse struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Current              struct {
		Time        string  `json:"time"`
		Temperature float64 `json:"temperature_2m"`
		FeelsLike   float64 `json:"apparent_temperature"`
		Humidity    int32   `json:"relative_humidity_2m"`
		WindSpeed   float64 `json:"wind_speed_10m"`
		WeatherCode int     `json:"weather_code"`
	} `json:"current"`
}

type forecastResponse struct {
	Latitude             float64 `json:"latitude"`
	Longitude            float64 `json:"longitude"`
	Timezone             string  `json:"timezone"`
	TimezoneAbbreviation string  `json:"timezone_abbreviation"`
	Daily                struct {
		Time                     []string  `json:"time"`
		TempMin                  []float64 `json:"temperature_2m_min"`
		TempMax                  []float64 `json:"temperature_2m_max"`
		PrecipitationProbability []int32   `json:"precipitation_probability_max"`
		WeatherCode              []int     `json:"weather_code"`
	} `json:"daily"`
}

func (c *WeatherClient) GetCurrentWeather(ctx context.Context, coordinates domain.Coordinates) (*domain.CurrentWeather, error) {
	params := url.Values{}
	params.Set("latitude", strconv.FormatFloat(coordinates.Latitude, 'f', -1, 64))
	params.Set("longitude", strconv.FormatFloat(coordinates.Longitude, 'f', -1, 64))
	params.Set("current", strings.Join([]string{
		"temperature_2m",
		"apparent_temperature",
		"relative_humidity_2m",
		"wind_speed_10m",
		"weather_code",
	}, ","))
	params.Set("timezone", "auto")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, weatherAPIURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather api returned status %d", resp.StatusCode)
	}

	var data currentWeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &domain.CurrentWeather{
		Coordinates: domain.Coordinates{
			Latitude:  data.Latitude,
			Longitude: data.Longitude,
		},
		Temperature: data.Current.Temperature,
		FeelsLike:   data.Current.FeelsLike,
		Humidity:    data.Current.Humidity,
		WindSpeed:   data.Current.WindSpeed,
		Condition:   weatherCondition(data.Current.WeatherCode),
		Time:        data.Current.Time,
		Timezone:    data.Timezone,
	}, nil
}

func (c *WeatherClient) GetForecast(ctx context.Context, coordinates domain.Coordinates, days int) ([]domain.ForecastDay, error) {
	params := url.Values{}
	params.Set("latitude", strconv.FormatFloat(coordinates.Latitude, 'f', -1, 64))
	params.Set("longitude", strconv.FormatFloat(coordinates.Longitude, 'f', -1, 64))
	params.Set("forecast_days", strconv.Itoa(days))
	params.Set("daily", strings.Join([]string{
		"temperature_2m_min",
		"temperature_2m_max",
		"precipitation_probability_max",
		"weather_code",
	}, ","))
	params.Set("timezone", "auto")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, weatherAPIURL+"?"+params.Encode(), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather api returned status %d", resp.StatusCode)
	}

	var data forecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	daysCount := len(data.Daily.Time)
	result := make([]domain.ForecastDay, 0, daysCount)

	for i := 0; i < daysCount; i++ {
		result = append(result, domain.ForecastDay{
			Date:                     data.Daily.Time[i],
			TempMin:                  data.Daily.TempMin[i],
			TempMax:                  data.Daily.TempMax[i],
			PrecipitationProbability: data.Daily.PrecipitationProbability[i],
			Condition:                weatherCondition(data.Daily.WeatherCode[i]),
		})
	}

	return result, nil
}

func weatherCondition(code int) domain.WeatherCondition {
	switch code {
	case 0:
		return domain.ConditionClear
	case 1:
		return domain.ConditionMostlyClear
	case 2:
		return domain.ConditionPartlyCloudy
	case 3:
		return domain.ConditionOvercast
	case 45, 48:
		return domain.ConditionFog
	case 51, 53, 55:
		return domain.ConditionDrizzle
	case 56, 57:
		return domain.ConditionFreezingDrizzle
	case 61, 63, 65:
		return domain.ConditionRain
	case 66, 67:
		return domain.ConditionFreezingRain
	case 71, 73, 75:
		return domain.ConditionSnow
	case 77:
		return domain.ConditionSnowGrains
	case 80, 81, 82:
		return domain.ConditionRainShowers
	case 85, 86:
		return domain.ConditionSnowShowers
	case 95:
		return domain.ConditionThunderstorm
	case 96, 99:
		return domain.ConditionThunderstormHail
	default:
		return domain.ConditionUnknown
	}
}
