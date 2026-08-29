package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/wxvn/grpc-mesh/internal/weather/domain"
)

var elevationAPIURL = "https://api.open-meteo.com/v1/elevation"

type ElevationClient struct {
	client *http.Client
}

func NewElevationClient(client *http.Client) *ElevationClient {
	return &ElevationClient{client: client}
}

type elevationResponse struct {
	Elevation []float64 `json:"elevation"`
}

func (c *ElevationClient) GetElevation(ctx context.Context, coordinates domain.Coordinates) (float64, error) {
	params := url.Values{}
	params.Set("latitude", strconv.FormatFloat(coordinates.Latitude, 'f', -1, 64))
	params.Set("longitude", strconv.FormatFloat(coordinates.Longitude, 'f', -1, 64))

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		elevationAPIURL+"?"+params.Encode(),
		nil,
	)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("elevation api returned status %d", resp.StatusCode)
	}

	var data elevationResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	if len(data.Elevation) == 0 {
		return 0, fmt.Errorf("elevation not found")
	}

	return data.Elevation[0], nil
}
