package clients

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

const (
	geocodingAPIURL = "https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=ru&format=json"
)

type GeocodingResponse struct {
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Geocoding struct {
	httpClient *http.Client
}

func NewGeocoding(httpClient *http.Client) *Geocoding {
	return &Geocoding{
		httpClient: httpClient,
	}
}

func (g *Geocoding) GetCoordinates(city string) (GeocodingResponse, error) {
	resp, err := g.httpClient.Get(
		fmt.Sprintf(geocodingAPIURL, city),
	)
	if err != nil {
		slog.Error(err.Error())
		return GeocodingResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("status code %d", resp.StatusCode)
		slog.Error(err.Error())
		return GeocodingResponse{}, fmt.Errorf("status code %d", resp.StatusCode)
	}

	var geoResp struct {
		Results []GeocodingResponse `json:"results"`
	}

	err = json.NewDecoder(resp.Body).Decode(&geoResp)
	if err != nil {
		slog.Error(err.Error())
		return GeocodingResponse{}, err
	}

	return geoResp.Results[0], nil
}
