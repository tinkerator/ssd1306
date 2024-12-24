// Program clock demonstrates text rendering on the ssd1306 display.
// It displays the time to the current second.
package main

import (
	"flag"
	"image"
	"image/color"
	"log"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"zappem.net/pub/io/device/ssd1306"
	"zappem.net/pub/io/i2c"
)

func main() {
	flag.Parse()

	dev, err := ssd1306.NewDevice(i2c.BusFile(1))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer dev.Close()

	const width = 128
	const height = 64
	const fSize = 16
	const skip = 8

	f, err := opentype.Parse(gomono.TTF)
	if err != nil {
		log.Fatalf("failed to parse gomono font: %v", err)
	}
	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    fSize,
		DPI:     72,
		Hinting: font.HintingNone,
	})
	if err != nil {
		log.Fatalf("failed to open new face: %v", err)
	}
	im := image.NewGray(image.Rect(0, 0, width, height))
	d := font.Drawer{
		Dst:  im,
		Src:  image.White,
		Face: face,
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	var then string
	for {
		now := <-ticker.C
		when := now.Format(time.DateTime)
		if when != then {
			lines := strings.Split(when, " ")
			for i, line := range lines {
				d.Dot = fixed.P(10, (height-skip)/2+i*(fSize+skip))
				d.DrawString(line)
			}
			dev.Display(im, func(col color.Color) bool {
				return color.GrayModel.Convert(col).(color.Gray).Y > 0
			})
			// clear the display
			for i := range im.Pix {
				im.Pix[i] = 0
			}
		}
		then = when
	}
}
