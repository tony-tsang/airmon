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
	SPI_CHUNK_SIZE = 4096
)

const (
	PSR = 0x00
	PWR = 0x01
	POF = 0x02
	PFS = 0x03
	PON = 0x04
	DTM1 = 0x10
	DSP = 0x11
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
	gpioInit bool
}

func (d *Display) Width() uint16 {
	return Width
}

func (d *Display) Height() uint16 {
	return Height
}

func (d *Display) Init(spiBus spi.Conn) {
	
	d.buffer = make([]byte, bufSz, bufSz)
 	d.spiBus = spiBus

}

func (d *Display) setup() {

	if !d.gpioInit {

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

		d.gpioInit = true
	}

	d.reset.Out(gpio.Low)
	time.Sleep(100 * time.Millisecond)
	d.reset.Out(gpio.High)

	d.busyWait(1000)

	cols, rows, resolution_setting := Width, Height, 0b10

	d.sendCommand(TRES, []byte{
		byte(cols >> 8),
		byte(cols & 0x0F),
		byte(rows >> 8),
		byte(rows & 0x0F),
	})

	d.sendCommand(PSR, []byte{
		byte(resolution_setting << 6 | 0b101111),
		0x08,
	})

	d.sendCommand(PWR, []byte{
		(0x6 << 3) | (0x01 << 2) | (0x01 << 1) | (0x01),
		0x00,
		0x23,
		0x23,
	})

	d.sendCommand(PLL, []byte{
		0x3C,
	})

	d.sendCommand(TSE, []byte{
		0x00,
	})
	d.sendCommand(CDI, []byte{
		0x1 << 5 | 0x17,
	})
	d.sendCommand(TCON, []byte{
		0x22,
	})
	d.sendCommand(DAM, []byte{
		0x00,
	})
	d.sendCommand(PWS, []byte{
		0xAA,
	})
	d.sendCommand(PFS, []byte{
		0x00,
	})

}

func (d *Display) SetPixel(x int, y int, color byte) {

	pos := int(Width / 2) * y + int(x / 2)
	
	if x & 1 == 0 {
		// set high bits
		d.buffer[pos] = (d.buffer[pos] & 0x0F) | ((color & 0x0F) << 4)
	} else {
		// set low bits
		d.buffer[pos] = (d.buffer[pos] & 0xF0) | (color & 0x0F)
	}

	// log.Printf("d.buffer[%v] = %v", pos, d.buffer[pos])
}

func (d *Display) Fill() {
	for y := 0; y < 400; y++ {
		for x := 0; x < 640; x++ {
			d.SetPixel(x, y, 6)
		}
	}
}

func (d *Display) busyWait(milliseconds int) {
	duration := time.Duration(milliseconds) * time.Millisecond

	log.Printf("start busy wait")

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

	log.Printf("end busy wait")
}

func (d *Display) sendCommand(cmd byte, data []byte) {

	buf := []byte{ cmd }
	log.Printf("cmd = 0x%X", cmd)
	d.spiWrite(SPI_COMMAND, buf)

	if data != nil {
		// log.Printf("data = %v", data)
		d.spiWrite(SPI_DATA, data)
	}
}

func (d *Display) UpdateScreen() {

	d.setup()

	log.Printf("sending DTM1")
	d.sendCommand(DTM1, d.buffer)
	d.busyWait(200)

	// log.Printf("sending DSP")
	// d.sendCommand(DSP, nil)
	// d.busyWait(200)

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

func (d *Display) spiWrite(dc gpio.Level, data []byte) {
	
	d.cs.Out(gpio.Low)

	d.dc.Out(dc)

	i := 0

	for {

		var end int

		if (i + SPI_CHUNK_SIZE) > len(data) {
			end = len(data)
		} else {
			end = i + SPI_CHUNK_SIZE
		}

		// log.Printf("spi TX data: %v", data[i:end])
		d.spiBus.Tx(data[i:end], nil)

		i += SPI_CHUNK_SIZE
		if i > len(data) {
			break
		}
	}

	d.cs.Out(gpio.High)	
}
