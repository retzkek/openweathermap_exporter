# Openweather exporter for prometheus


Exporter for openweather API 


## Quickstart

Create an API key from https://openweathermap.org/.

Install dependancies with `go get` and then build the binary.

```
go get -d -v
go build
OWM_LOCATION=LONDON,UK  OWM_API_KEY=apikey ./openweathermap_exporter
```

Then add the scraper in prometheus

```
scrape_configs:
  - job_name: 'weather'

    # Scrape is configured for free usage.
    scrape_interval: 60s

    # Port is not yet configurable
    static_configs:
      - targets: ['localhost:2112']
```



## With Docker

The image is a multistage image, just launch as usual :

```
docker build -t ows .
docker run --rm -e OWM_LOCATION=LONDON,UK  -e OWM_API_KEY=apikey -p 2112:2112 ows
```

