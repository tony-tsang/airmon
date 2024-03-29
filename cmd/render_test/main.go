package main

import (
    "github.com/fogleman/gg"
    "github.com/makeworld-the-better-one/dither/v2"
    "github.com/tony-tsang/airmon/assets"
    "github.com/tony-tsang/airmon/internal/pkg/dps310"
    "github.com/tony-tsang/airmon/internal/pkg/hko"
    "github.com/tony-tsang/airmon/internal/pkg/uc8159"
    "image/color"
    "image/png"
    "log"
    "periph.io/x/conn/v3/i2c/i2creg"
    "periph.io/x/conn/v3/physic"
    "periph.io/x/conn/v3/spi"
    "periph.io/x/conn/v3/spi/spireg"
    "periph.io/x/host/v3"
    "sync"
    "time"
)

type AllData struct {
    TempPressure dps310.TempPressure
    HKOData      hko.HKOData
}

func main() {

    state, err := host.Init()
    if err != nil {
        log.Fatalf("failed to initialize periph: %v", err)
    }

    log.Printf("Using drivers:\n")
    for _, driver := range state.Loaded {
        log.Printf("- %s", driver)
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

    spiChannel, err := spiBus.Connect(physic.MegaHertz, spi.Mode3|spi.NoCS, 8)

    if _, ok := spiChannel.(spi.Pins); ok {
        //log.Printf("  CLK : %s", p.CLK())
        //log.Printf("  MOSI: %s", p.MOSI())
        //log.Printf("  MISO: %s", p.MISO())
        //log.Printf("  CS  : %s", p.CS())
    }

    d := new(uc8159.Display)
    d.Init(spiChannel)

    hkodatachannel := make(chan hko.HKOData)
    go hko.DoLoop(hkodatachannel, 5*time.Minute)

    pressurechannel := make(chan dps310.TempPressure)
    go dps310.DoLoop(i2cBus, pressurechannel, 10*time.Second)

    context := gg.NewContext(640, 400)

    // fileData, err := os.ReadFile("NotoSansTC-Light.otf")
    // if err != nil {
    // 	log.Fatalf("Unable to load font file: %v", err)
    // }

    //f, err := opentype.Parse(assets.FontData)
    //if err != nil {
    //    log.Fatalf("Unable to parse font: %v", err)
    //}

    drawMutex := sync.Mutex{}

    context.SetColor(color.White)
    context.Clear()

    allData := AllData{}

    go func() {

        for {
            drawMutex.Lock()

            pngFile, err := assets.IconFS.Open("icons/01n.png")
            if err != nil {
                log.Fatal("Error opening 01n.png")
            }
            svgImage, err := png.Decode(pngFile)
            if err != nil {
                log.Fatal("Error decoding png")
            }

            /*
               face, err := opentype.NewFace(f, &opentype.FaceOptions{
                   Size:    24,
                   DPI:     72,
                   Hinting: font.HintingNone,
               })
               if err != nil {
                   log.Fatalf("NewFace: %v", err)
               }*/

            /*
               context.SetFontFace(face)
               context.SetColor(color.Black)
               context.DrawString(allData.HKOData.GeneralSituation, 0, 60)

               face, err = opentype.NewFace(f, &opentype.FaceOptions{
                   Size:    36,
                   DPI:     72,
                   Hinting: font.HintingNone,
               })
               if err != nil {
                   log.Fatalf("NewFace: %v", err)
               }
               text := fmt.Sprintf("%0.2f hPa %0.2f ℃", allData.TempPressure.Pressure, allData.TempPressure.Temp)
               context.SetFontFace(face)
               w, h := context.MeasureString(text)
               log.Printf("w=%v, h=%v", w, h)
               context.SetColor(color.RGBA{0, 0, 255, 0xFF})
               context.DrawRectangle(0, 0, w, h)
               context.SetColor(color.Black)
               context.DrawStringWrapped(text, 0, 32, 0, 0, 640, 2, gg.AlignLeft)

               image := context.Image()
            */
            palette := color.Palette{
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
            dithered := ditherer.Dither(svgImage)

            for y := 0; y < 400; y++ {
                for x := 0; x < 640; x++ {
                    pixel := dithered.At(x, y)
                    var index int
                    _, _, _, alpha := pixel.RGBA()
                    if alpha == 0 {
                        index = 1
                    } else {
                        index = matchPalette(palette, pixel)
                    }
                    d.SetPixel(x, y, uc8159.Color(index))
                }
            }

            d.UpdateScreen()
            drawMutex.Unlock()
            context.SetColor(color.White)
            context.Clear()
            time.Sleep(2 * time.Minute)
        }
    }()

    for {

        select {
        case hkodata := <-hkodatachannel:
            log.Printf("hkoData %v", hkodata)
            drawMutex.Lock()
            allData.HKOData = hkodata

            drawMutex.Unlock()
        case tempData := <-pressurechannel:
            log.Printf("tempPressure %v", tempData)
            drawMutex.Lock()
            allData.TempPressure = tempData
            drawMutex.Unlock()
        default:
            time.Sleep(5 * time.Second)
        }

    }
}

func matchPalette(palette color.Palette, pixel color.Color) int {
    return palette.Index(pixel)
}
