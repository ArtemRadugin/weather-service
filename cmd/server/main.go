package main

import (
	"encoding/json"
	"fmt"
	"github.com/ArtemRadugin/weather-service/internal/client/http/geocoding"
	"github.com/ArtemRadugin/weather-service/internal/client/http/open_meteo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
	"log"
	"net/http"
	"sync"
	"time"
)

const (
	httpPort = ":3000"
)

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	geocodingClient := geocoding.NewClient(httpClient)
	openMeteoClient := open_meteo.NewClient(httpClient)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		fmt.Printf("received request for city: %s\n", city)

		geoResp, err := geocodingClient.GetCoordinates(city)
		if err != nil {
			log.Println(err)
			return
		}

		opMetResp, err := openMeteoClient.GetTemperature(geoResp.Latitude, geoResp.Longitude)
		if err != nil {
			log.Println(err)
			return
		}

		row, err := json.Marshal(opMetResp)
		if err != nil {
			log.Println(err)
		}

		_, err = w.Write(row)
		if err != nil {
			log.Println(err)
		}
	})

	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}

	jobs, err := initJobs(s)
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		fmt.Println("starting http server on port", httpPort)
		err := http.ListenAndServe(httpPort, r)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		fmt.Printf("starting job: %v", jobs[0].ID())
		s.Start()
	}()

	wg.Wait()
}

func initJobs(scheduler gocron.Scheduler) ([]gocron.Job, error) {
	j, err := scheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				fmt.Println("cron job executed")
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil
}
