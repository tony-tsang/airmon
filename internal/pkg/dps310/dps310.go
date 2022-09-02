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
    PRSB2       = 0x00
    TMPB2       = 0x03
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
    overSampleScaleFactor []int
    pressureScale         int
    temperatureScale      int
    seaLevelPressure      float64
}

func New(i2cBus i2c.Bus) *DPS310 {
    d := new(DPS310)
    d.dev = &i2c.Dev{Addr: sensorAddr, Bus: i2cBus}
    d.overSampleScaleFactor = []int{
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

func (d *DPS310) Reset() {
    outBuf := []byte{RESET, 0x89}
    err := d.dev.Tx(outBuf, nil)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }
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

func (d *DPS310) readCalibration() {

}
