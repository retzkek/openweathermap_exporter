package main

import (
	"context"
	"log"
	"net/http"
	"time"

	owm "github.com/briandowns/openweathermap"
	"github.com/caarlos0/env"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config stores the parameters used to fetch the data
type Config struct {
	pollingInterval time.Duration
	requestTimeout  time.Duration
	APIKey          string `env:"OWM_API_KEY"` // APIKey delivered by Openweathermap
	Location        string `env:"OWM_LOCATION" envDefault:"Lille,FR"`
	Language        string `env:"OWM_LANGUAGE" envDefault:"fr"`
	Duration        int    `env:"OWM_DURATION" envDefault:"5"`
}

func loadMetrics(ctx context.Context, location string) <-chan error {
	errC := make(chan error)
	go func() {
		c := time.Tick(cfg.pollingInterval)
		for {
			select {
			case <-ctx.Done():
				return // returning not to leak the goroutine
			case <-c:
				client := &http.Client{
					Timeout: cfg.requestTimeout,
				}

				w, err := owm.NewCurrent("C", cfg.Language, cfg.APIKey, owm.WithHttpClient(client))
				if err != nil {
					errC <- err
					continue
				}

				err = w.CurrentByName(location)
				if err != nil {
					errC <- err
					continue
				}

				temp.WithLabelValues(location).Set(w.Main.Temp)

				pressure.WithLabelValues(location).Set(w.Main.Pressure)

				humidity.WithLabelValues(location).Set(float64(w.Main.Humidity))

				wind.WithLabelValues(location).Set(w.Wind.Speed)

				clouds.WithLabelValues(location).Set(float64(w.Clouds.All))

				rain.WithLabelValues(location).Set(w.Rain.ThreeH)

				var scraped_weather = w.Weather[0].Description
				if scraped_weather == last_weather {
					weather.WithLabelValues(location, scraped_weather).Set(1)
				} else {
					weather.WithLabelValues(location, scraped_weather).Set(1)
					weather.WithLabelValues(location, last_weather).Set(0)
					last_weather = scraped_weather
				}
				log.Println(w.Weather[0].Description)
				log.Println("scraping OK for ", location)
			}
		}
	}()
	return errC
}

var (
	cfg = Config{
		pollingInterval: 5 * time.Second,
		requestTimeout:  1 * time.Second,
	}
	temp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "temperature_celsius",
		Help:      "Temperature in °C",
	}, []string{"location"})

	pressure = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "pressure_hpa",
		Help:      "Atmospheric pressure in hPa",
	}, []string{"location"})

	humidity = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "humidity_percent",
		Help:      "Humidity in Percent",
	}, []string{"location"})

	wind = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "wind_mps",
		Help:      "Wind speed in m/s",
	}, []string{"location"})

	clouds = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "cloudiness_percent",
		Help:      "Cloudiness in Percent",
	}, []string{"location"})
	rain = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "rain",
		Help:      "Rain contents 3h",
	}, []string{"location"})
	weather = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "openweathermap",
		Name:      "weather",
		Help:      "The weather label.",
	}, []string{"location", "weather"})

	last_weather = ""
)

func main() {

	env.Parse(&cfg)
	prometheus.Register(temp)
	prometheus.Register(pressure)
	prometheus.Register(humidity)
	prometheus.Register(wind)
	prometheus.Register(clouds)
	prometheus.Register(weather)

	errC := loadMetrics(context.TODO(), cfg.Location)
	go func() {
		for err := range errC {
			log.Println(err)
		}
	}()
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
