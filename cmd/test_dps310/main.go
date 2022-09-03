package test_dps310

import (
    "fmt"
    "github.com/tony-tsang/airmon/internal/pkg/dps310"
    "log"
    "periph.io/x/conn/v3/driver/driverreg"
    "periph.io/x/conn/v3/i2c/i2creg"
    "periph.io/x/host/v3"
    "time"
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

    i2cBus, err := i2creg.Open("")
    if err != nil {
        log.Fatalf("failed to open I2C: %v", err)
    }

    defer i2cBus.Close()

    sensor := dps310.New(i2cBus)
    sensor.Init()

    for {
        pressure := sensor.GetPressure()
        fmt.Printf("pressure: %v", pressure)
        time.Sleep(1 * time.Second)
    }
}
