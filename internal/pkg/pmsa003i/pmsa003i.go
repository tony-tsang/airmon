package pmsa003i

import (
    "encoding/binary"
    "log"
    "time"

    "periph.io/x/conn/v3/i2c"
)

type PMSensorValue struct {
    FrameLength    uint16
    PM10std        uint16
    PM25std        uint16
    PM100std       uint16
    PM10env        uint16
    PM25env        uint16
    PM100env       uint16
    Particles03um  uint16
    Particles05um  uint16
    Particles10um  uint16
    Particles25um  uint16
    Particles50um  uint16
    Particles100um uint16
    Version        byte
    ErrorCode      byte
}

const (
    PmSensorAddr uint16 = 0x0012
)

type PMSA003I struct {
    dev *i2c.Dev
}

func New(i2cBus i2c.Bus) *PMSA003I {
    d := new(PMSA003I)
    d.dev = &i2c.Dev{Addr: PmSensorAddr, Bus: i2cBus}
    return d
}

func (d *PMSA003I) Read() PMSensorValue {
    inBuffer := make([]byte, 32)
    err := d.dev.Tx(nil, inBuffer)
    if err != nil {
        log.Printf("Error reading from Temp sensor %d\n", err)
    }

    var checkSum uint16 = 0

    values := PMSensorValue{
        binary.BigEndian.Uint16(inBuffer[2:4]),
        binary.BigEndian.Uint16(inBuffer[4:6]),
        binary.BigEndian.Uint16(inBuffer[6:8]),
        binary.BigEndian.Uint16(inBuffer[8:10]),
        binary.BigEndian.Uint16(inBuffer[10:12]),
        binary.BigEndian.Uint16(inBuffer[12:14]),
        binary.BigEndian.Uint16(inBuffer[14:16]),
        binary.BigEndian.Uint16(inBuffer[16:18]),
        binary.BigEndian.Uint16(inBuffer[18:20]),
        binary.BigEndian.Uint16(inBuffer[20:22]),
        binary.BigEndian.Uint16(inBuffer[22:24]),
        binary.BigEndian.Uint16(inBuffer[24:26]),
        binary.BigEndian.Uint16(inBuffer[26:28]),
        inBuffer[28],
        inBuffer[29],
    }

    checkSum += 0x42
    checkSum += 0x4d

    for i := 2; i < 30; i++ {
        checkSum += uint16(inBuffer[i])
    }

    devCheckSum := binary.BigEndian.Uint16(inBuffer[30:32])

    if devCheckSum != checkSum {
        log.Printf("Checksum mismatch dev 0x%x != calc 0x%x", devCheckSum, checkSum)
    }

    return values
}

func DoLoop(i2cBus i2c.Bus, channel chan PMSensorValue, sleep time.Duration) {

    d := New(i2cBus)

    for {
        values := d.Read()
        channel <- values
        time.Sleep(sleep)
    }

}
