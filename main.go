package main

import (
    htu31 "airmon/htu31"
    "airmon/metrics"
    pmsa003i "airmon/pmsa003i"
    "flag"
    "github.com/d2r2/go-logger"
    "log"
    "time"
)

func main() {

    logger.ChangePackageLogLevel("i2c", logger.InfoLevel)

    var sleepInterval int
    var listenAddress string

    flag.IntVar(&sleepInterval,"interval", 10, "sensor read interval in seconds")
    flag.StringVar(&listenAddress, "listen", ":8080", "listen address for prometheus metrics")
    flag.Parse()

    sleepDuration := time.Duration(sleepInterval) * time.Second

    tempHumidityChannel := make(chan htu31.TempHumidity)
    go htu31.DoLoop(1, tempHumidityChannel, sleepDuration)

    pmValueChannel := make(chan pmsa003i.PMSensorValue)
    go pmsa003i.DoLoop(1, pmValueChannel, sleepDuration)

    go metrics.StartServer(listenAddress)

    for {
        select {
            case tempHumidity := <-tempHumidityChannel:
                log.Printf("Temperature %.2f, humidity %.2f\n", tempHumidity.Temp, tempHumidity.Humidity)
                metrics.TemperatureMetric.Set(tempHumidity.Temp)
                metrics.HumidityMetric.Set(tempHumidity.Humidity)

            case pmValue := <-pmValueChannel:
                log.Printf("PM1.0 %d PM2.5 %d PM10 %d", pmValue.PM10std, pmValue.PM25std, pmValue.PM100std)
                metrics.PM10StdMetric.Set(float64(pmValue.PM10std))
                metrics.PM25StdMetric.Set(float64(pmValue.PM25std))
                metrics.PM100StdMetric.Set(float64(pmValue.PM100std))
            default:
                time.Sleep(1 * time.Second)
        }
    }
}