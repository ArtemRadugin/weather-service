package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ArtemRadugin/weather-service/internal/clients"
	"github.com/ArtemRadugin/weather-service/internal/clients/http/geocoding"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const (
	httpPort = ":3000"
	city     = "moscow"
)

type Reading struct {
	Name        string    `db:"name"`
	Timestamp   time.Time `db:"timestamp"`
	Temperature float64   `db:"temperature"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	ctx := context.Background()

	connStr := "postgresql://postgres:qwerty@localhost:8080/postgres"

	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		cityName := chi.URLParam(r, "city")
		fmt.Printf("received request for city: %s\n", cityName)

		var reading Reading
		err := conn.QueryRow(
			ctx,
			"SELECT name, timestamp, temperature FROM reading WHERE name = $1 ORDER BY timestamp DESC LIMIT 1",
			city,
		).Scan(&reading.Name, &reading.Timestamp, &reading.Temperature)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("not found"))
				return
			}
		}

		row, err := json.Marshal(reading)
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

	jobs, err := initJobs(ctx, s, conn)
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

func initJobs(ctx context.Context, scheduler gocron.Scheduler, conn *pgx.Conn) ([]gocron.Job, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	geocodingClient := geocoding.NewClient(httpClient)
	openMeteoClient := clients.NewClient(httpClient)

	j, err := scheduler.NewJob(
		gocron.DurationJob(
			1*time.Minute,
		),
		gocron.NewTask(
			func() {
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

				timestamp, err := time.Parse("2006-01-02T15:04", opMetResp.Current.Time)
				if err != nil {
					log.Println(err)
					return
				}

				_, err = conn.Exec(
					ctx,
					"INSERT INTO reading (name, timestamp, temperature) VALUES ($1, $2, $3)",
					city, timestamp, opMetResp.Current.Temperature2m)
				if err != nil {
					log.Println(err)
					return
				}

				fmt.Printf("%v updated data for city: %s\n", time.Now(), city)
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil
}
