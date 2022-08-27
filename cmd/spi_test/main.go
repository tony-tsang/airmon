package main

import (
    "log"
    "image"
    "os"
    "image/color"
    _ "image/jpeg"

    "periph.io/x/host/v3"
    "periph.io/x/conn/v3/physic"
    "periph.io/x/conn/v3/spi"
    "periph.io/x/conn/v3/spi/spireg"
    "periph.io/x/conn/v3/i2c/i2creg"
    "github.com/makeworld-the-better-one/dither/v2"

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

    imageFile, err := os.Open("sample.jpg")

    defer imageFile.Close()

    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    imageData, _, err := image.Decode(imageFile)

    if err != nil {
        log.Fatalf("Error: %v", err)
    }

    palette := []color.Color{
        color.RGBA{0, 0, 0, 0xFF},
        color.RGBA{255, 255, 255, 0xFF},
        color.RGBA{0, 255, 0, 0xFF},
        color.RGBA{0, 0, 255, 0xFF},
        color.RGBA{255, 0, 0, 0xFF},
        color.RGBA{255, 255, 0, 0xFF},
        color.RGBA{255, 140, 0, 0xFF},
        color.RGBA{255, 255, 255, 0xFF},
    }

    ditherer := dither.NewDitherer(palette)

    ditherer.Matrix = dither.FloydSteinberg

    img := ditherer.Dither(imageData)

    // log.Printf("%v", img)

    for y := 0; y < 400; y++ {
        for x := 0; x < 640; x++ {
            color := img.At(x, y)
            index := matchPalette(palette, color)
            d.SetPixel(x, y, uc8159.Color(index))
        }
    }
    
    //d.Fill(uc8159.WHITE)

    d.UpdateScreen()
}

func matchPalette(palette []color.Color, pixel color.Color) int {
    for i, color := range palette {
        if color == pixel {
            return i
        }
    }

    return 0
}