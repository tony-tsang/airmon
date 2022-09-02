package htu31

import (
    "encoding/binary"
    "log"
    "time"

    "periph.io/x/conn/v3/i2c"
)

const (
    sensorAddr uint16 = 0x0040

    SoftReset     = 0x1E
    ReadSerial    = 0x0A
    Conversion    = 0x5E
    ReadTempHumid = 0x00
    HeaterOff     = 0x02

    ConversionPause = 25 * time.Millisecond // 25 milliseconds
)

type TempHumidity struct {
    Temp     float64
    Humidity float64
}

type Device struct {
    dev *i2c.Dev
}

func New(i2cBus i2c.Bus) *Device {
    d := new(Device)
    d.dev = &i2c.Dev{Addr: sensorAddr, Bus: i2cBus}
    return d
}

func (d *Device) SoftReset() {
    outBuffer := []byte{SoftReset}
    err := d.dev.Tx(outBuffer, nil)
    if err != nil {
        log.Printf("Error writing reset to Temp sensor %d\n", err)
    }
}

func (d *Device) HeaterOff() {
    outBuffer := []byte{HeaterOff}
    err := d.dev.Tx(outBuffer, nil)
    if err != nil {
        log.Printf("Error writing HeaterOff to Temp sensor %d\n", err)
    }
}

func (d *Device) ReadSerial() uint32 {
    inBuffer := make([]byte, 6)
    outBuffer := []byte{ReadSerial}
    err := d.dev.Tx(outBuffer, inBuffer)
    if err != nil {
        log.Printf("Error reading from Temp sensor %d\n", err)
    }

    serial := binary.BigEndian.Uint32(inBuffer)
    return serial
}

func (d *Device) Conversion() {
    outBuffer := []byte{Conversion}
    err := d.dev.Tx(outBuffer, nil)
    if err != nil {
        log.Printf("Error writing to Temp sensor %d\n", err)
    }
}

func (d *Device) ReadTempHumid() (float64, float64) {

    outBuffer := []byte{ReadTempHumid}
    inBuffer := make([]byte, 6)
    err := d.dev.Tx(outBuffer, inBuffer)
    if err != nil {
        log.Printf("Error writing to Temp sensor %d\n", err)
    }

    temperatureRaw := binary.BigEndian.Uint16(inBuffer[0:2])
    temperatureCRC := inBuffer[2]

    crcValue := crc(uint32(temperatureRaw))

    var temperature float64

    if crcValue != temperatureCRC {
        log.Printf("CRC incorrect for temperature, dev 0x%x != calc 0x%x", temperatureCRC, crcValue)
    }

    //log.Printf("temperature Raw = 0x%x %d", temperatureRaw, temperatureRaw)

    temperature = -40 + 165*float64(temperatureRaw)/(2<<15-1)

    humidityRaw := binary.BigEndian.Uint16(inBuffer[3:5])
    humidityCRC := inBuffer[5]

    crcValue = crc(uint32(humidityRaw))

    if crcValue != humidityCRC {
        log.Printf("CRC incorrect for humidity, dev 0x%x != calc 0x%x", humidityCRC, crcValue)
    }

    humidity := 100 * float64(humidityRaw) / (2<<15 - 1)

    return temperature, humidity
}

func DoLoop(i2cBus i2c.Bus, channel chan TempHumidity, sleep time.Duration) {

    d := New(i2cBus)

    d.SoftReset()
    time.Sleep(500 * time.Millisecond)
    d.HeaterOff()
    time.Sleep(1 * time.Second)

    serial := d.ReadSerial()
    log.Printf("serial %d\n", serial)

    sleep = sleep - ConversionPause

    for {
        d.Conversion()

        time.Sleep(ConversionPause)

        temperature, humidity := d.ReadTempHumid()

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
        if result&msb != 0 {
            result = ((result ^ polynom) & mask) | (result & ^mask)
        }

        msb >>= 1
        mask >>= 1
        polynom >>= 1

    }

    return uint8(result)
}
