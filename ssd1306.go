// Package ssd1306 connects to a SSD1306 device over i2c and can
// render images on it.
package ssd1306

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
	"sync"

	"zappem.net/pub/io/i2c"
)

// addr is the device address.
const addr = 0x3c

// these constants are from https://github.com/supprot/ArducamSSD1306/blob/master/ArducamSSD1306.h

type CommandByte byte

const (
	SetLowColumn                     CommandByte = 0x00
	ExternalVCC                                  = 0x01
	SwitchCapVCC                                 = 0x02
	SetHighColumn                                = 0x10
	MemoryMode                                   = 0x20
	ColumnAddr                                   = 0x21
	PageAddr                                     = 0x22
	RightHorizontalScroll                        = 0x26
	LeftHorizontal_scroll                        = 0x27
	VerticalAndRightHorizontalScroll             = 0x29
	VerticalAndLeftHorizontalScroll              = 0x2A
	DeactivateScroll                             = 0x2E
	ActivateScroll                               = 0x2F
	SetStartLine                                 = 0x40
	SetContrast                                  = 0x81
	ChargePump                                   = 0x8D
	SegRemap                                     = 0xA0
	SegRemapHigh                                 = 0xA1
	SetVerticalScrollArea                        = 0xA3
	DisplayAllOnResume                           = 0xA4
	DisplayAllOn                                 = 0xA5
	NormalDisplay                                = 0xA6
	InvertDisplay                                = 0xA7
	SetMultiplex                                 = 0xA8
	DisplayOff                                   = 0xAE
	DisplayOn                                    = 0xAF
	ComScanInc                                   = 0xC0
	ComScanDec                                   = 0xC8
	SetDisplayOffset                             = 0xD3
	SetDisplayClockDiv                           = 0xD5
	SetPrecharge                                 = 0xD9
	SetComPins                                   = 0xDA
	SetVComDetect                                = 0xDB
)

// Device holds a connection to the SSD1306 device.
type Device struct {
	mu sync.Mutex
	c  io.ReadWriteCloser
}

// command executes one of the byte commands. Note the arguments are
// sent as if they are separate byte commands.
func (dev *Device) command(cmd byte) error {
	n, err := dev.c.Write(append([]byte{0, cmd}))
	if err != nil {
		return err
	}
	if n != 2 {
		return fmt.Errorf("bad write: sent=%d, wanted=2", n)
	}
	return nil
}

// Cmd sends a command to the SSD1306 device. Care when using
// this. For simple operation, this more direct access should not be
// needed.
func (dev *Device) Cmd(cmd CommandByte, args ...byte) error {
	dev.mu.Lock()
	defer dev.mu.Unlock()

	if err := dev.command(byte(cmd)); err != nil {
		return err
	}
	for i, b := range args {
		if err := dev.command(b); err != nil {
			return fmt.Errorf("sending arg[%d]=0x%2x: %v", i, b, err)
		}
	}
	return nil
}

// Reset reinitializes an open device. It runs the start up sequence.
func (dev *Device) Reset() {
	dev.mu.Lock()
	defer dev.mu.Unlock()

	var setup = []CommandByte{
		DisplayOff,
		SetDisplayClockDiv, 0x80,
		SetMultiplex, 0x3f,
		SetDisplayOffset, 0x00,
		SetStartLine,
		ChargePump, 0x14,
		MemoryMode, 0x00,
		SegRemapHigh,
		ComScanDec,
		SetComPins, 0x12,
		SetContrast, 0xcf,
		SetPrecharge, 0xf1,
		SetVComDetect,
		SetStartLine,
		DisplayAllOnResume,
		NormalDisplay,
		DisplayOn,
	}
	for _, c := range setup {
		dev.command(byte(c))
	}
}

// NewDevice opens a SSD1306 device via the busfile device.
func NewDevice(busfile string) (*Device, error) {
	c, err := i2c.NewConn(busfile, addr, false, binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	dev := &Device{
		c: c,
	}
	dev.Reset()
	return dev, nil
}

// Close closes the device.
func (dev *Device) Close() error {
	return dev.c.Close()
}

// Display renders an image.Image on the device, using the callback
// function pixel() to determine which pixel color.Color values should
// be represented with a dot.
func (dev *Device) Display(im image.Image, pixel func(color.Color) bool) {
	dev.mu.Lock()
	defer dev.mu.Unlock()

	dev.command(byte(ColumnAddr))
	dev.command(0x00)
	dev.command(0x7f)
	dev.command(byte(PageAddr))
	dev.command(0x00)
	dev.command(0x07)

	for i := 0; i < 64; i++ {
		data := []byte{byte(SetStartLine)}
		c, r := (i%8)*16, (i/8)*8
		var datum byte
		for o := 0; o < 128; o++ {
			datum <<= 1
			dc := c + (o >> 3)
			dr := r + 7 - (o & 7)
			if pixel(im.At(dc, dr)) {
				datum |= 1
			}
			if o&7 == 7 {
				data = append(data, datum)
			}
		}
		dev.c.Write(data)
	}
}
