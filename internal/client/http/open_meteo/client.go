package open_meteo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Current struct {
		Time          string  `json:"time"`
		Temperature2m float64 `json:"temperature_2m"`
	}
}

type client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *client {
	return &client{
		httpClient: httpClient,
	}
}

func (c *client) GetTemperature(lat, long float64) (Response, error) {
	// Implementation goes here
	resp, err := c.httpClient.Get(
		fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%f&longitude=%f&current=temperature_2m",
			lat,
			long,
		),
	)
	if err != nil {
		return Response{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Response{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var tempResponse Response
	err = json.NewDecoder(resp.Body).Decode(&tempResponse)
	if err != nil {
		return Response{}, err
	}

	return tempResponse, nil
}
