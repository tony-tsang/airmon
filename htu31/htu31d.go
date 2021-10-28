package htu31

import (
    "encoding/binary"
    "github.com/d2r2/go-i2c"
    "log"
    "time"
)


const (
    TempSensorAddr uint8 = 0x40

    SoftReset     = 0x1E
    ReadSerial    = 0x0A
    Conversion    = 0x5E
    ReadTempHumid = 0x00

    ConversionPause = 25 * time.Millisecond  // 25 milliseconds
)

type TempHumidity struct {
    Temp     float64
    Humidity float64
}


func DoLoop(i2cBus int, channel chan TempHumidity, sleep time.Duration) {

    tempSensor, err := i2c.NewI2C(TempSensorAddr, i2cBus)
    if err != nil { log.Fatal(err) }

    defer tempSensor.Close()

    outBuffer := []byte {SoftReset}
    inBuffer := make([]byte, 6)

    _, err = tempSensor.WriteBytes(outBuffer)
    if err != nil { log.Printf("Error writing to Temp sensor %d\n", err) }

    outBuffer = []byte {ReadSerial}
    _, err = tempSensor.WriteBytes(outBuffer)
    if err != nil { log.Printf("Error writing to Temp sensor %d\n", err) }

    _, err = tempSensor.ReadBytes(inBuffer)
    if err != nil { log.Printf("Error reading from Temp sensor %d\n", err) }

    serial := binary.BigEndian.Uint32(inBuffer)
    log.Printf("serial %d\n", serial)

    sleep = sleep - ConversionPause

    for {

        outBuffer = []byte {Conversion}
        _, err = tempSensor.WriteBytes(outBuffer)
        if err != nil { log.Printf("Error writing to Temp sensor %d\n", err) }

        time.Sleep(ConversionPause)

        outBuffer = []byte {ReadTempHumid}
        _, err = tempSensor.WriteBytes(outBuffer)
        if err != nil { log.Printf("Error writing to Temp sensor %d\n", err) }

        _, err = tempSensor.ReadBytes(inBuffer)
        if err != nil { log.Printf("Error reading from Temp sensor %d\n", err) }

        temperatureRaw := binary.BigEndian.Uint16(inBuffer[0:2])
        temperatureCRC := inBuffer[2]

        crcValue := crc(uint32(temperatureRaw))

        var temperature float64

        if crcValue != temperatureCRC {
            log.Printf("CRC incorrect for temperature, dev 0x%x != calc 0x%x", temperatureCRC, crcValue)
            time.Sleep(sleep)
            continue
        }

        temperature = -40 + 165*float64(temperatureRaw)/(2<<15-1)

        humidityRaw := binary.BigEndian.Uint16(inBuffer[3:5])
        humidityCRC := inBuffer[5]

        crcValue = crc(uint32(humidityRaw))

        if crcValue != humidityCRC {
            log.Printf("CRC incorrect for humidity, dev 0x%x != calc 0x%x", humidityCRC, crcValue)
            time.Sleep(sleep)
            continue
        }

        humidity := 100 * float64(humidityRaw) / (2<<15 - 1)

        channel <- TempHumidity{temperature, humidity}

        time.Sleep(sleep)
    }

}


func crc(input uint32) uint8 {

    var polynom uint32 = 0x988000
    var msb uint32 = 0x800000
    var mask uint32 = 0xFF8000

    result := input << 8

    for msb != 0x80 {
        if result & msb != 0 {
            result = ((result ^ polynom) & mask) | (result & ^mask)
        }

        msb >>= 1
        mask >>= 1
        polynom >>= 1

    }

    return uint8(result)
}