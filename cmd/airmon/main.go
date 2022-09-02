package main

import (
    "flag"
    "log"
    "time"

    "periph.io/x/conn/v3/driver/driverreg"
    "periph.io/x/conn/v3/i2c/i2creg"
    "periph.io/x/conn/v3/physic"
    "periph.io/x/conn/v3/spi"
    "periph.io/x/conn/v3/spi/spireg"
    "periph.io/x/host/v3"

    "github.com/tony-tsang/airmon/internal/pkg/htu31"
    "github.com/tony-tsang/airmon/internal/pkg/metrics"
    "github.com/tony-tsang/airmon/internal/pkg/pmsa003i"
)

func main() {

    _, err := host.Init()
    if err != nil {
        log.Fatalf("failed to initialize periph: %v", err)
    }

    _, err = driverreg.Init()
    if err != nil {
        log.Fatalf("failed to initialize periph: %v", err)
    }

    spiBus, err := spireg.Open("")
    if err != nil {
        log.Fatalf("failed to open SPI: %v", err)
    }

    i2cBus, err := i2creg.Open("")
    if err != nil {
        log.Fatalf("failed to open I2C: %v", err)
    }

    defer spiBus.Close()
    defer i2cBus.Close()

    _, err = spiBus.Connect(physic.MegaHertz, spi.Mode3, 8)

    var sleepInterval int
    var listenAddress string

    flag.IntVar(&sleepInterval, "interval", 10, "sensor read interval in seconds")
    flag.StringVar(&listenAddress, "listen", ":8080", "listen address for prometheus metrics")
    flag.Parse()

    sleepDuration := time.Duration(sleepInterval) * time.Second

    tempHumidityChannel := make(chan htu31.TempHumidity)
    go htu31.DoLoop(i2cBus, tempHumidityChannel, sleepDuration)

    pmValueChannel := make(chan pmsa003i.PMSensorValue)
    go pmsa003i.DoLoop(i2cBus, pmValueChannel, sleepDuration)

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
            metrics.PM10Concentration.Set(float64(pmValue.PM10env))
            metrics.PM25Concentration.Set(float64(pmValue.PM25env))
            metrics.PM100Concentration.Set(float64(pmValue.PM100env))
        default:
            time.Sleep(1 * time.Second)
        }
    }
}
