package clients

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

const (
	openMeteoAPIURL = "https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m"
)

type OpenMeteoResponse struct {
	Current struct {
		Time          string  `json:"time"`
		Temperature2m float64 `json:"temperature_2m"`
	}
}

type OpenMeteo struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *OpenMeteo {
	return &OpenMeteo{
		httpClient: httpClient,
	}
}

func (o *OpenMeteo) GetTemperature(lat, long float64) (OpenMeteoResponse, error) {
	// Implementation goes here
	resp, err := o.httpClient.Get(
		fmt.Sprintf(openMeteoAPIURL, lat, long),
	)
	if err != nil {
		slog.Error(err.Error())
		return OpenMeteoResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("status code: %d", resp.StatusCode)
		slog.Error(err.Error())
		return OpenMeteoResponse{}, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	var response OpenMeteoResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		slog.Error(err.Error())
		return OpenMeteoResponse{}, err
	}

	return response, nil
}
