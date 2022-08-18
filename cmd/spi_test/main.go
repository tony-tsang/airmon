package main

import (
    // "flag"
    "log"
    // "time"
    "periph.io/x/host/v3"
    "periph.io/x/conn/v3/physic"
    "periph.io/x/conn/v3/driver/driverreg"
    "periph.io/x/conn/v3/spi"
    "periph.io/x/conn/v3/spi/spireg"
    // "periph.io/x/conn/v3/i2c"
    "periph.io/x/conn/v3/i2c/i2creg"
    // "periph.io/x/conn/v3/gpio"
    // "periph.io/x/conn/v3/gpio/gpioreg"

    // "github.com/tony-tsang/airmon/internal/pkg/htu31"
    // "github.com/tony-tsang/airmon/internal/pkg/metrics"
    // "github.com/tony-tsang/airmon/internal/pkg/pmsa003i"
	"github.com/tony-tsang/airmon/internal/pkg/uc8159"
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

    spi_bus, err := spireg.Open("")
    if err != nil {
        log.Fatalf("failed to open SPI: %v", err)
    }

    i2c_bus, err := i2creg.Open("")
    if err != nil {
        log.Fatalf("failed to open I2C: %v", err)
    }
    
    defer spi_bus.Close()
    defer i2c_bus.Close()
    
    spi_channel, err := spi_bus.Connect(physic.MegaHertz, spi.Mode3, 8)

	var d uc8159.Display
	d.Init(spi_channel)
}