package pmsa003i

import (
    "encoding/binary"
    "github.com/d2r2/go-i2c"
    "log"
    "time"
)

type PMSensorValue struct {
    FrameLength uint16
    PM10std uint16
    PM25std uint16
    PM100std uint16
    PM10env uint16
    PM25env uint16
    PM100env uint16
    Particles03um uint16
    Particles05um uint16
    Particles10um uint16
    Particles25um uint16
    Particles50um uint16
    Particles100um uint16
    Version byte
    ErrorCode byte
}

const (
    PmSensorAddr   uint8 = 0x12
)

func DoLoop(i2cBus int, channel chan PMSensorValue, sleep time.Duration) {

    sensor, err := i2c.NewI2C(PmSensorAddr, i2cBus)
    if err != nil { log.Fatal(err) }

    defer sensor.Close()

    inBuffer := make([]byte, 32)

    for {

        _, err = sensor.ReadBytes(inBuffer)
        if err != nil { log.Printf("Error reading from Temp sensor %d\n", err) }

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

        channel <- values

        time.Sleep(sleep)
    }

}