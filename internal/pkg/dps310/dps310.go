package dps310

import (
    "log"
    "periph.io/x/conn/v3/i2c"
    "time"
)

const (
    sensorAddr = 0x77
)

const (
    PSRB2       = 0x00
    PSRB1       = 0x01
    PSRB0       = 0x02
    TMPB2       = 0x03
    TMPB1       = 0x04
    TMPB0       = 0x05
    PRSCFG      = 0x06
    TMPCFG      = 0x07
    MEASCFG     = 0x08
    RESET       = 0x0C
    PRODREVID   = 0x0D
    TMPCOEFSRCE = 0x28
)

const (
    COEF_RDY   = 7
    SENSOR_RDY = 6
    TMP_RDY    = 5
    PRS_RDY    = 4
)

const (
    IDLE                             = 0
    PRESSURE_MEASURE                 = 0b001
    TEMPERATURE_MEASURE              = 0b010
    CONTINUOUS_PRESSURE_MEASURE      = 0b101
    CONTINUOUS_TEMPERATURE_MEASURE   = 0b110
    CONTINUOUS_TEMP_PRESSURE_MEASURE = 0b111
)

type DPS310 struct {
    dev                   *i2c.Dev
    overSampleScaleFactor []int32
    pressureScale         int32
    temperatureScale      int32
    seaLevelPressure      float64

    c0  int32
    c1  int32
    c00 int32
    c10 int32
    c01 int32
    c11 int32
    c20 int32
    c21 int32
    c30 int32
}

func New(i2cBus i2c.Bus) *DPS310 {
    d := new(DPS310)
    d.dev = &i2c.Dev{Addr: sensorAddr, Bus: i2cBus}
    d.overSampleScaleFactor = []int32{
        524288,
        1572864,
        3670016,
        7864320,
        253952,
        516096,
        1040384,
        2088960,
    }

    d.pressureScale = d.overSampleScaleFactor[6]
    d.temperatureScale = d.overSampleScaleFactor[6]
    d.seaLevelPressure = 1013.25

    return d
}

func (d *DPS310) Init() {
    d.Reset()
    d.WaitTemperatureReady()
    d.WaitPressureReady()
}

func (d *DPS310) correctTemp() {
    outBuf := []byte{0x0E, 0xA5}
    err := d.dev.Tx(outBuf, nil)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    outBuf = []byte{0x0F, 0x96}
    err = d.dev.Tx(outBuf, nil)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    outBuf = []byte{0x62, 0x02}
    err = d.dev.Tx(outBuf, nil)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    outBuf = []byte{0x0E, 0x00}
    err = d.dev.Tx(outBuf, nil)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    outBuf = []byte{0x0F, 0x00}
    err = d.dev.Tx(outBuf, nil)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    outBuf = []byte{TMPB2, 0x00}
    err = d.dev.Tx(outBuf, outBuf)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }
}

func (d *DPS310) Reset() {
    outBuf := []byte{RESET, 0x89}
    err := d.dev.Tx(outBuf, nil)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    time.Sleep(10 * time.Millisecond)
    for !d.sensorReady() {
        time.Sleep(1 * time.Millisecond)
    }

    d.correctTemp()
    d.readCalibration()

    config := d.calibCoeffTempSrcBit()
    d.setTempMeasurementSrcBit(config)
}

func (d *DPS310) getCfgRegister(bit uint8) uint8 {
    outBuf := []byte{MEASCFG, 0x0}
    err := d.dev.Tx(outBuf, outBuf)
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }

    return (outBuf[1] >> bit) & 0x1
}

func (d *DPS310) WaitTemperatureReady() {
    for d.getCfgRegister(TMP_RDY) == 0 {
        time.Sleep(1 * time.Millisecond)
    }
}

func (d *DPS310) WaitPressureReady() {
    for d.getCfgRegister(PRS_RDY) == 0 {
        time.Sleep(1 * time.Millisecond)
    }
}

func twosComplement(val uint32, bits int) int32 {
    if val & (1 << (bits - 1)) {
        val -= 1 << bits
    }

    return int32(val)
}

func (d *DPS310) readCalibration() {
    for !d.coefficientsReady() {
        time.Sleep(1 * time.Millisecond)
    }

    coeffs := make([]uint32, 18)

    for offset := 0; offset < 18; offset++ {
        buffer := make([]byte, 2)
        buffer[0] = 0x10 + byte(offset)

        err := d.dev.Tx(buffer, buffer)

        if err != nil {
            log.Printf("Error reading / writing to sensor %d\n", err)
        }
        coeffs[offset] = int32(buffer[1])
    }

    x := (coeffs[0] << 4) | ((coeffs[1] >> 4) & 0x0F)
    d.c0 = twosComplement(x, 12)

    d.c1 = twosComplement(((coeffs[1]&0x0F)<<8)|coeffs[2], 12)

    x = (coeffs[3] << 12) | (coeffs[4] << 4) | ((coeffs[5] >> 4) & 0x0F)
    d.c00 = twosComplement(x, 20)

    x = ((coeffs[5] & 0x0F) << 16) | (coeffs[6] << 8) | coeffs[7]
    d.c10 = twosComplement(x, 20)

    d.c01 = twosComplement((coeffs[8]<<8)|coeffs[9], 16)
    d.c11 = twosComplement((coeffs[10]<<8)|coeffs[11], 16)
    d.c20 = twosComplement((coeffs[12]<<8)|coeffs[13], 16)
    d.c21 = twosComplement((coeffs[14]<<8)|coeffs[15], 16)
    d.c30 = twosComplement((coeffs[16]<<8)|coeffs[17], 16)
}

func (d *DPS310) sensorReady() bool {
    return d.getCfgRegister(SENSOR_RDY) == 1
}

func (d *DPS310) coefficientsReady() bool {
    return d.getCfgRegister(COEF_RDY) == 1
}

func (d *DPS310) calibCoeffTempSrcBit() byte {

    buffer := []byte{TMPCOEFSRCE, 0x00}
    err := d.dev.Tx(buffer, buffer)

    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }

    val := buffer[1] & (0x01 << 7)

    return val
}

func (d *DPS310) setTempMeasurementSrcBit(val byte) {
    buffer := []byte{TMPCFG, val}
    err := d.dev.Tx(buffer, nil)

    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }
}

func (d *DPS310) rawTemperature() uint32 {
    buffer := []byte{TMPB2, 0x0}
    err := d.dev.Tx(buffer, buffer)
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }
    tmpB2 := buffer[1]

    buffer = []byte{TMPB1, 0x0}
    err = d.dev.Tx(buffer, buffer)
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }
    tmpB1 := buffer[1]

    buffer = []byte{TMPB0, 0x0}
    err = d.dev.Tx(buffer, buffer)
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }
    tmpB0 := buffer[1]

    rawTemperature := uint32(tmpB2)<<24 | uint32(tmpB1)<<16 | uint32(tmpB0)

    return rawTemperature
}

func (d *DPS310) rawPressure() uint32 {
    buffer := []byte{PSRB2, 0x0}
    err := d.dev.Tx(buffer, buffer)
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }
    psrB2 := buffer[1]

    buffer = []byte{PSRB1, 0x0}
    err = d.dev.Tx(buffer, buffer)
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }
    psrB1 := buffer[1]

    buffer = []byte{PSRB0, 0x0}
    err = d.dev.Tx(buffer, buffer)
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }
    psrB0 := buffer[1]

    rawPressure := uint32(psrB2)<<24 | uint32(psrB1)<<16 | uint32(psrB0)

    return rawPressure
}

func (d *DPS310) GetPressure() int32 {
    tempReading := d.rawTemperature()
    rawTemperature := twosComplement(tempReading, 24)

    pressureReading := d.rawPressure()
    rawPressure := twosComplement(pressureReading, 24)

    scaledRawTemp := rawTemperature / d.temperatureScale
    scaledRawPressure := rawPressure / d.pressureScale

    presCalc := d.c00 +
        scaledRawPressure +
        (d.c10+scaledRawPressure)*(d.c20+scaledRawPressure*d.c30) +
        scaledRawTemp +
        (d.c01+scaledRawPressure)*(d.c11+scaledRawPressure*d.c21)

    finalPressure := presCalc / 100
    return finalPressure
}

func (d *DPS310) GetTemperature() float64 {
    rawTemperature := int32(d.rawTemperature())
    scaledRawTemp := float64(rawTemperature) / float64(d.temperatureScale)
    temperature := scaledRawTemp*float64(d.c1) + float64(d.c0)/2.0
    return temperature
}
