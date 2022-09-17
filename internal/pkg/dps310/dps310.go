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
    CFGREG      = 0x09
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

    d.init()

    return d
}

func (d *DPS310) setCfgRegisterBit(bit byte, val bool) {

    outBuf := []byte{CFGREG}
    inBuf := make([]byte, 1)

    err := d.dev.Tx(outBuf, inBuf)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    log.Printf("bit = %v, val = %v", bit, val)
    log.Printf("inBuf = %08b", inBuf[0])
    if val {
        inBuf[0] |= byte(1 << bit)
    } else {
        inBuf[0] &= 0xFF ^ byte(1<<bit)
    }

    log.Printf("inBuf = %08b", inBuf[0])
    outBuf = []byte{CFGREG, inBuf[0]}
    err = d.dev.Tx(outBuf, nil)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }
}

func (d *DPS310) setPressureCfg() {
    log.Printf("setPressureCfg")
    val := byte(0b0110)
    outBuf := []byte{PRSCFG, val}
    err := d.dev.Tx(outBuf, nil)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    d.setCfgRegisterBit(2, true)
}

func (d *DPS310) setTemperatureCfg() {

    val := byte(0b0110)
    outBuf := []byte{TMPCFG, val}
    currentVal := make([]byte, 1)
    err := d.dev.Tx(outBuf, currentVal)

    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    currentVal[0] = currentVal[0] | val
    outBuf = []byte{TMPCFG, currentVal[0]}
    err = d.dev.Tx(outBuf, nil)

    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    d.setCfgRegisterBit(3, true)
}

func (d *DPS310) setCfgBits(cmd, offset, length, val byte) {

    outBuf := []byte{cmd}
    inBuf := make([]byte, 1)
    err := d.dev.Tx(outBuf, inBuf)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    bitmask := byte(0)
    for i := offset; i < offset+length; i++ {
        bitmask |= 1 << i
    }
    log.Printf("cmd = %02x, offset %v, length %v, bitmask = %08b, val = %08b", cmd, offset, length, bitmask, val)

    currentVal := inBuf[0]

    log.Printf("currentVal = %08b", currentVal)

    currentVal |= (currentVal & ^bitmask) | val<<offset

    log.Printf("currentVal = %08b", currentVal)

    outBuf = []byte{cmd, currentVal}
    err = d.dev.Tx(outBuf, nil)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }
}

func (d *DPS310) setMode() {
    log.Printf("setMode")
    d.setCfgBits(MEASCFG, 0, 3, CONTINUOUS_TEMP_PRESSURE_MEASURE)
}

func (d *DPS310) init() {
    log.Printf("dps310 init")
    d.readProductRevID()
    d.reset()

    d.setPressureCfg()
    d.setTemperatureCfg()
    d.setMode()

    d.waitTemperatureReady()
    d.waitPressureReady()
}

func (d *DPS310) readProductRevID() {
    outBuf := []byte{PRODREVID}
    inBuf := make([]byte, 1)

    err := d.dev.Tx(outBuf, inBuf)
    if err != nil {
        log.Printf("Error writing to sensor %d\n", err)
    }

    log.Printf("prod revision id = 0x%x", inBuf[0])
}

func (d *DPS310) correctTemp() {

    log.Printf("CorrectTemp")

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

    _ = d.rawTemperature()
}

func (d *DPS310) reset() {
    log.Printf("dps310 reset")
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
    // log.Printf("calib coefficient temp src bit: %08b", config)
    d.setTempMeasurementSrcBit(config)
}

func (d *DPS310) getCfgRegister(bit uint8) uint8 {
    outBuf := []byte{MEASCFG}
    inBuf := make([]byte, 1)
    err := d.dev.Tx(outBuf, inBuf)
    //log.Printf("inBuf=%08b", inBuf[0])
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }

    // log.Printf("cfg register: %08b", inBuf[0])
    return (inBuf[0] >> bit) & 0x1
}

func (d *DPS310) waitTemperatureReady() {
    for d.getCfgRegister(TMP_RDY) == 0 {
        time.Sleep(1 * time.Millisecond)
    }
}

func (d *DPS310) waitPressureReady() {
    for d.getCfgRegister(PRS_RDY) == 0 {
        time.Sleep(1 * time.Millisecond)
    }
}

func twosComplement(val uint32, bits int) int32 {
    if val&(1<<(bits-1)) != 0 {
        val -= 1 << bits
    }

    return int32(val)
}

func (d *DPS310) readCalibration() {
    for !d.coefficientsReady() {
        time.Sleep(1 * time.Millisecond)
    }

    coeffs := make([]byte, 18)

    for offset := 0; offset < 18; offset++ {
        buffer := []byte{0x10 + byte(offset)}
        data := make([]byte, 1)

        err := d.dev.Tx(buffer, data)

        if err != nil {
            log.Printf("Error reading / writing to sensor %d\n", err)
        }
        coeffs[offset] = data[0]
    }

    x := uint32(coeffs[0])<<4 | uint32((coeffs[1]>>4)&0x0F)
    d.c0 = twosComplement(x, 12)

    d.c1 = twosComplement(
        (uint32(coeffs[1])&0x0F)<<8|
            uint32(coeffs[2]), 12)

    x = uint32(coeffs[3])<<12 | uint32(coeffs[4])<<4 | uint32((coeffs[5]>>4)&0x0F)
    d.c00 = twosComplement(x, 20)

    x = (uint32(coeffs[5])&0x0F)<<16 | uint32(coeffs[6])<<8 | uint32(coeffs[7])
    d.c10 = twosComplement(x, 20)

    d.c01 = twosComplement(
        uint32(coeffs[8])<<8|
            uint32(coeffs[9]), 16)
    d.c11 = twosComplement(
        uint32(coeffs[10])<<8|
            uint32(coeffs[11]), 16)
    d.c20 = twosComplement(
        uint32(coeffs[12])<<8|
            uint32(coeffs[13]), 16)
    d.c21 = twosComplement(
        uint32(coeffs[14])<<8|
            uint32(coeffs[15]), 16)
    d.c30 = twosComplement(
        uint32(coeffs[16])<<8|
            uint32(coeffs[17]), 16)

    log.Printf("readCalibration: d= %v %v %v %v %v %v %v %v %v", d.c0, d.c1, d.c00, d.c10, d.c01, d.c11, d.c20, d.c21, d.c30)
}

func (d *DPS310) sensorReady() bool {
    value := d.getCfgRegister(SENSOR_RDY)
    return value == 1
}

func (d *DPS310) coefficientsReady() bool {
    value := d.getCfgRegister(COEF_RDY)
    return value == 1
}

func (d *DPS310) calibCoeffTempSrcBit() byte {

    buffer := []byte{TMPCOEFSRCE}
    inBuf := make([]byte, 1)
    err := d.dev.Tx(buffer, inBuf)

    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }

    val := (inBuf[0] & (0x01 << 7)) >> 7

    return val
}

func (d *DPS310) setTempMeasurementSrcBit(val byte) {

    buffer := []byte{TMPCFG}
    currentVal := make([]byte, 1)
    err := d.dev.Tx(buffer, currentVal)

    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }

    currentVal[0] = currentVal[0] | (val << 7)

    buffer = []byte{TMPCFG, currentVal[0]}
    err = d.dev.Tx(buffer, nil)

    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }
}

func (d *DPS310) rawTemperature() uint32 {

    buffer := []byte{TMPB2}
    inBuf := make([]byte, 3)
    err := d.dev.Tx(buffer, inBuf)
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }

    rawTemperature := uint32(inBuf[0])<<16 | uint32(inBuf[1])<<8 | uint32(inBuf[2])

    return rawTemperature
}

func (d *DPS310) rawPressure() uint32 {
    buffer := []byte{PSRB2}
    inBuf := make([]byte, 1)
    err := d.dev.Tx(buffer, inBuf)
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }
    psrB2 := inBuf[0]

    buffer = []byte{PSRB1}
    err = d.dev.Tx(buffer, inBuf)
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }
    psrB1 := inBuf[0]

    buffer = []byte{PSRB0}
    err = d.dev.Tx(buffer, inBuf)
    if err != nil {
        log.Printf("Error reading / writing to sensor %d\n", err)
    }
    psrB0 := inBuf[0]

    rawPressure := uint32(psrB2)<<16 | uint32(psrB1)<<8 | uint32(psrB0)

    return rawPressure
}

func (d *DPS310) GetPressure() float64 {
    tempReading := d.rawTemperature()
    rawTemperature := twosComplement(tempReading, 24)

    pressureReading := d.rawPressure()
    rawPressure := twosComplement(pressureReading, 24)

    log.Printf("raw_pressure=%v", rawPressure)

    scaledRawTemp := float64(rawTemperature) / float64(d.temperatureScale)
    scaledRawPressure := float64(rawPressure) / float64(d.pressureScale)

    presCalc := float64(d.c00) +
        scaledRawPressure*
            (float64(d.c10)+scaledRawPressure*(float64(d.c20)+scaledRawPressure*float64(d.c30))) +
        scaledRawTemp*
            (float64(d.c01)+scaledRawPressure*(float64(d.c11)+scaledRawPressure*float64(d.c21)))

    finalPressure := presCalc / 100.0
    return finalPressure
}

func (d *DPS310) GetTemperature() float64 {
    rawTemperature := int32(d.rawTemperature())
    scaledRawTemp := float64(rawTemperature) / float64(d.temperatureScale)
    temperature := scaledRawTemp*float64(d.c1) + float64(d.c0)/2.0
    return temperature
}
