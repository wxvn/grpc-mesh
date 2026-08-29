package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/wxvn/grpc-mesh/internal/weather/domain"
)

var geocodingAPIURL = "https://geocoding-api.open-meteo.com/v1/search"

type GeocodingClient struct {
	client *http.Client
}

func NewGeocodingClient(client *http.Client) *GeocodingClient {
	return &GeocodingClient{client: client}
}

type geocodingResponse struct {
	Results []struct {
		Name      string  `json:"name"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	} `json:"results"`
}

func (c *GeocodingClient) GetCoordinates(ctx context.Context, city string) (domain.Coordinates, error) {
	params := url.Values{}
	params.Set("name", city)
	params.Set("count", "1")
	params.Set("language", "ru")
	params.Set("format", "json")

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		geocodingAPIURL+"?"+params.Encode(),
		nil,
	)
	if err != nil {
		return domain.Coordinates{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return domain.Coordinates{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return domain.Coordinates{}, fmt.Errorf("geocoding api returned status %d", resp.StatusCode)
	}

	var data geocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return domain.Coordinates{}, err
	}

	if len(data.Results) == 0 {
		return domain.Coordinates{}, fmt.Errorf("city %q not found", city)
	}

	return domain.Coordinates{
		Latitude:  data.Results[0].Latitude,
		Longitude: data.Results[0].Longitude,
	}, nil
}
