package uc8159

import (
	"time"
	"log"
	
	"periph.io/x/conn/v3/spi"
	"periph.io/x/conn/v3/gpio"
    "periph.io/x/conn/v3/gpio/gpioreg"
)

const Width = 640
const Height = 400

const Spi_Chunk_Size = 4096

const (
	RESET_PIN = "GPIO27"
	BUSY_PIN = "GPIO17"
	DC_PIN = "GPIO22"
	CS0_PIN = "GPIO8"
)

const (
	SPI_COMMAND = gpio.Low
	SPI_DATA = gpio.High
)

const (
	PSR = 0x00
	PWR = 0x01
	POF = 0x02
	PFS = 0x03
	PON = 0x04
	DTM1 = 0x10
	DRF = 0x12
	IPC = 0x13
	PLL = 0x30
	TSE = 0x41
	TSW = 0x42
	TSR = 0x43
	CDI = 0x50
	LPD = 0x51
	TCON = 0x60
	TRES = 0x61
	DAM = 0x65
	REV = 0x70
	FLG = 0x71
	AMV = 0x80
	VV = 0x81
	VDCS = 0x82
	PWS = 0xE3
	TSSET = 0xE5
)

const bufSz = Width / 2 * Height

type Display struct {
	spiBus spi.Conn
	buffer []byte
	reset gpio.PinIO
	busy gpio.PinIO
	dc gpio.PinIO
	cs gpio.PinIO
}

func (d *Display) Width() uint16 {
	return Width
}

func (d *Display) Height() uint16 {
	return Height
}

func (d *Display) Init(spiBus spi.Conn) {
	
	d.buffer = make([]byte, bufSz)
	d.spiBus = spiBus

	d.reset = gpioreg.ByName(RESET_PIN)
	if d.reset == nil {
		log.Fatalf("Unable to find %s", RESET_PIN)
	}
	d.reset.Out(gpio.High)

	d.dc = gpioreg.ByName(DC_PIN)
	if d.dc == nil {
		log.Fatalf("Unable to find %s", DC_PIN)
	}
	d.dc.Out(gpio.Low)

	d.cs = gpioreg.ByName(CS0_PIN)
	if d.cs == nil {
		log.Fatalf("Unable to find %s", CS0_PIN)
	}
	d.cs.Out(gpio.High)

	d.busy = gpioreg.ByName(BUSY_PIN)
	if d.busy == nil {
		log.Fatalf("Unable to find %s", BUSY_PIN)
	}
	d.busy.In(gpio.PullDown, gpio.NoEdge)

	d.Reset()
}

func (d *Display) Reset() {
	d.reset.Out(gpio.Low)
	time.Sleep(100 * time.Millisecond)
	d.reset.Out(gpio.High)

}

func (d *Display) Fill() {
	for i := range d.buffer {
		d.buffer[i] = 0x2 & 0x07
	}
}

func (d *Display) busyWait(milliseconds int) {
	duration := time.Duration(milliseconds) * time.Millisecond

	if (d.busy.Read() == gpio.High) {
		// device is not busy
		time.Sleep(duration)
	} else {
		start := time.Now()
		var current time.Time
		var elapsed time.Duration

		for d.busy.Read() == gpio.Low {

			time.Sleep(10 * time.Millisecond)

			current = time.Now()
			elapsed = current.Sub(start)
			if elapsed >= duration {
				log.Printf("Wait duration expired, buy busy flag still Low")
				return
			}
		}
	}

}

func (d *Display) sendCommand(cmd byte, data *[]byte) {

	buf := make([]byte, 1)
	buf[0] = cmd
	log.Printf("cmd = %v", cmd)
	d.spiWrite(SPI_COMMAND, &buf)

	if data != nil {
		d.spiWrite(SPI_DATA, data)
	}
}

func (d *Display) UpdateScreen() {
	log.Printf("sending DTM1")
	d.sendCommand(DTM1, &d.buffer)

	log.Printf("sending PON")
	d.sendCommand(PON, nil)
	d.busyWait(200)

	log.Printf("sending DRF")
	d.sendCommand(DRF, nil)
	d.busyWait(32000)

	log.Printf("sending POF")
	d.sendCommand(POF, nil)
	d.busyWait(200)
}

func (d *Display) spiWrite(dc gpio.Level, data *[]byte) {
	
	read := make([]byte, 0)

	// log.Printf("d = %v", d)
	log.Printf("CS low")
	d.cs.Out(gpio.Low)
	log.Printf("dc = %v", dc)
	d.dc.Out(dc)

	d.spiBus.Tx(*data, read)

	log.Printf("CS High")
	d.cs.Out(gpio.High)	
}