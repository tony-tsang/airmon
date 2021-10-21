package cmd

import (
    "encoding/binary"
    "github.com/d2r2/go-i2c"
    "log"
    "time"
)


const (
    I2cBus                 = 1
    PmSensorAddr   uint8 = 0x12
    TempSensorAddr uint8 = 0x40

    HTU31DSoftReset = 0x1E
    HTU31DReadSerial = 0x0A
    HTU31DConversion = 0x40
    HTU31DReadTempHumid = 0x00

    HTU31DConversionPause = 25 // 25 milliseconds
)

func do_loop() {

    tempSensor, err := i2c.NewI2C(TempSensorAddr, I2cBus)
    if err != nil { log.Fatal(err) }

    pmSensor, err := i2c.NewI2C(PmSensorAddr, I2cBus)
    if err != nil { log.Fatal(err) }

    defer tempSensor.Close()
    defer pmSensor.Close()

    outBuffer := []byte {HTU31DSoftReset}
    inBuffer := make([]byte, 16)

    tempSensor.WriteBytes(outBuffer)

    for {
        outBuffer = []byte {HTU31DConversion}
        n, err := tempSensor.WriteBytes(outBuffer)
        if err != nil { log.Printf("Error writing to temp sensor %d\n", err) }
        log.Printf("Written %d bytes\n", n)

        time.Sleep(HTU31DConversionPause * time.Millisecond)

        outBuffer = []byte {HTU31DReadTempHumid}
        n, err = tempSensor.WriteBytes(outBuffer)
        if err != nil { log.Printf("Error writing to temp sensor %d\n", err) }
        log.Printf("Written %d bytes\n", n)

        n, err = tempSensor.ReadBytes(inBuffer)
        if err != nil { log.Printf("Error reading from temp sensor %d\n", err) }

        log.Printf("Read %d bytes\n", n)

        temperatureByte := binary.BigEndian.Uint16(inBuffer[0:2])

        temperature := -40 + 165 * float32(temperatureByte) / (2<<16 - 1)

        log.Printf("temperature %.2f\n", temperature)

        time.Sleep(10 * time.Second)


    }

}