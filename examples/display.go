// Program display demonstrates displaying something on a SSD1306
// OLED device via i2c.
package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"log"
	"os"

	"zappem.net/pub/io/i2c"
	"zappem.net/pub/io/ssd1306"
)

var (
	pixel = flag.Int("pixel", 125, "pixel displayed when gray no smaller")
	x     = flag.Int("x", 19, "width of a rectangle (ignored if --image)")
	y     = flag.Int("y", 8, "height of a rectangle (ignored if --image)")
	img   = flag.String("image", "", "PNG file")
)

func gray(col color.Color) bool {
	return int(color.GrayModel.Convert(col).(color.Gray).Y) >= *pixel
}

func main() {
	flag.Parse()

	dev, err := ssd1306.NewDevice(i2c.BusFile(1))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer dev.Close()

	var im image.Image
	if *img != "" {
		f, err := os.Open(*img)
		if err != nil {
			log.Fatalf("failed to load image %q: %v", *img, err)
		}
		im, err = png.Decode(f)
		f.Close()
		if err != nil {
			log.Fatalf("unable to decode %q: %v", *img, err)
		}
	} else {
		im1 := image.NewRGBA(image.Rect(0, 0, 128, 64))
		green := color.RGBA{0, 255, 0, 255}
		draw.Draw(im1, image.Rect(0, 0, *x, *y), &image.Uniform{green}, image.ZP, draw.Src)
		im = im1
	}

	dev.Display(im, gray)
}
