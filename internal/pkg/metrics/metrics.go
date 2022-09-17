package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "log"
    "net/http"
)

var (
    TemperatureMetric = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "temperature",
        Help: "Current temperature",
    })

    HumidityMetric = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "humidity",
        Help: "Current humidity",
    })

    PM10StdMetric = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "pm10std",
        Help: "PM1.0 Standard",
    })

    PM25StdMetric = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "pm25std",
        Help: "PM2.5 Standard",
    })

    PM100StdMetric = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "pm100std",
        Help: "PM10 Standard",
    })

    PM10Concentration = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "pm10concentration",
        Help: "PM1.0 Concentration",
    })

    PM25Concentration = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "pm25concentration",
        Help: "PM2.5 Concentration",
    })

    PM100Concentration = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "pm100concentration",
        Help: "PM10 Concentration",
    })

    PressureMetric = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "pressure",
        Help: "Atmospheric pressure",
    })
)

func StartServer(addr string) {

    http.Handle("/metrics", promhttp.Handler())
    err := http.ListenAndServe(addr, nil)
    if err != nil {
        log.Fatal("Error listening socket, ", err)
    }

}
