package main

import (
	"flag"
	"github.com/gobuffalo/packr"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"image"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"
)

const (
	// FontProp Antiproportional for font size for impact font text
	FontProp = 10.0
	// Offset Offset of the text from the top and bottom
	Offset = 10
)

var (
	imgPath    = flag.String("i", "", "Image file to edit")
	outPath    = flag.String("o", "", "Output file")
	topText    = flag.String("t", "", "Top text")
	bottomText = flag.String("b", "", "Bottom text")
)

func initImpact(size float64) font.Face {
	fontDataPath := packr.NewBox(".")
	fontData, err := fontDataPath.Find("impact.ttf")
	if err != nil {
		log.Fatal(err)
	}

	font, err := truetype.Parse(fontData)
	if err != nil {
		log.Fatal(err)
	}

	var options truetype.Options
	options.Size = size

	face := truetype.NewFace(font, &options)
	return face
}

func drawImpactStr(drawer font.Drawer, s string, p fixed.Point26_6) {
	// Get normalized forwarding
	drawer.Dot = fixed.P(0, 0)
	bounds, advance := drawer.BoundString(s)
	height := bounds.Max.Y - bounds.Min.Y

	// Draw at center
	drawer.Dot = fixed.Point26_6{
		X: p.X - advance/2,
		Y: p.Y + height/2}
	drawer.DrawString(s)
}

func impacc(src string, dst string, top string, bottom string) {
	// Read in
	file, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}

	img, fmt, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	bounds := img.Bounds()
	width, height := bounds.Max.X-bounds.Min.X, bounds.Max.Y-bounds.Min.Y
	fontSize := float64(height) / FontProp

	// Read font and initialize face
	impact := initImpact(fontSize)

	// Edit image
	drawImg := image.NewRGBA(bounds)
	draw.Draw(drawImg, bounds, img, image.Point{0, 0}, draw.Src)

	drawer := font.Drawer{
		Face: impact,
		Src:  image.White,
		Dst:  drawImg,
		Dot:  fixed.P(0, 0)}

	// Top text
	drawImpactStr(
		drawer,
		strings.ToUpper(top),
		fixed.P(width/2, int(fontSize/2)+Offset))

	// Bottom text
	drawImpactStr(
		drawer,
		strings.ToUpper(bottom),
		fixed.P(width/2, height-int(fontSize/2)-Offset))

	// Write out
	out, err := os.Create(dst)
	if err != nil {
		log.Fatal(err)
	}

	// Write back in original format
	switch fmt {
	case "gif":
		var options gif.Options
		gif.Encode(out, drawer.Dst, &options)
	case "jpeg":
		var options jpeg.Options
		options.Quality = 100
		jpeg.Encode(out, drawer.Dst, &options)
	case "png":
		png.Encode(out, drawer.Dst)
	}

	defer out.Close()
}

func main() {
	flag.Parse()
	impacc(*imgPath, *outPath, *topText, *bottomText)
}
