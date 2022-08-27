package main

import (
    // "flag"
    "log"
    // "time"
    "periph.io/x/host/v3"
    "periph.io/x/conn/v3/physic"
    // "periph.io/x/conn/v3/driver/driverreg"
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
	state, err := host.Init()
    if err != nil {
		log.Fatalf("failed to initialize periph: %v", err)
	}

    log.Printf("Using drivers:\n")
    for _, driver := range state.Loaded {
        log.Printf("- %s", driver)
    }

    // _, err = driverreg.Init()
    // if err != nil {
    //     log.Fatalf("failed to initialize periph: %v", err)
    // }

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
    
    spi_channel, err := spi_bus.Connect(physic.MegaHertz, spi.Mode3 | spi.NoCS, 8)

    if p, ok := spi_channel.(spi.Pins); ok {
        log.Printf("  CLK : %s", p.CLK())
		log.Printf("  MOSI: %s", p.MOSI())
		log.Printf("  MISO: %s", p.MISO())
		log.Printf("  CS  : %s", p.CS())
    }

	d := new(uc8159.Display)
	d.Init(spi_channel)

    d.Fill()

    d.UpdateScreen()
}