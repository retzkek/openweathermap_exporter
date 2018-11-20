package main

import (
	owm "github.com/briandowns/openweathermap"
	"github.com/caarlos0/env"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"time"
)

type Config struct {
	ApiKey   string `env:"OWM_API_KEY"`
	Location string `env:"OWM_LOCATION" envDefault:"Lille,FR"`
	Duration int    `env:"OWM_DURATION" envDefault:5"`
}

func load_metrics(location string) {

	for {
		
		go func(location string) {
			w, err := owm.NewCurrent("C", "FR", cfg.ApiKey) // (internal - OpenWeatherMap reference for kelvin) with English output
			if err != nil {
				log.Fatalln(err)
			}

			w.CurrentByName(location)

			temp.WithLabelValues(location).Set(w.Main.Temp)

			pressure.WithLabelValues(location).Set(w.Main.Pressure)

			humidity.WithLabelValues(location).Set(float64(w.Main.Humidity))

			wind.WithLabelValues(location).Set(w.Wind.Speed)

			clouds.WithLabelValues(location).Set(float64(w.Clouds.All))
			log.Println("scraping OK for ", location)
		}(location)
		time.Sleep(60 * time.Second)
	}
}

var (
	cfg  = Config{}
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
)

func main() {

	env.Parse(&cfg)
	prometheus.Register(temp)
	prometheus.Register(pressure)
	prometheus.Register(humidity)
	prometheus.Register(wind)
	prometheus.Register(clouds)

	go load_metrics(cfg.Location)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
