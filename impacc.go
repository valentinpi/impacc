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
	FontProp = 16.0
	// Offset Offset of the text from the top and bottom
	Offset = 10
	// Outline Thickness of text outline
	Outline = 2
)

var (
	imgPath    = flag.String("i", "", "Image file to edit")
	outPath    = flag.String("o", "", "Output file")
	topText    = flag.String("t", "", "Top text")
	bottomText = flag.String("b", "", "Bottom text")
	fontSize   = 0.0
)

func initImpact() *truetype.Font {
	fontDataPath := packr.NewBox(".")
	fontData, err := fontDataPath.Find("impact.ttf")
	if err != nil {
		log.Fatal(err)
	}

	font, err := truetype.Parse(fontData)
	if err != nil {
		log.Fatal(err)
	}

	return font
}

func initFace(font *truetype.Font, size float64) font.Face {
	face := truetype.NewFace(font, &truetype.Options{
		Size: size})
	return face
}

// Does NOT preserve the drawer given
// Only requires the Dst field to be initialized with the output buffer
func drawImpactStr(impactFont *truetype.Font, drawer *font.Drawer, s string, p fixed.Point26_6) {
	impact := initFace(impactFont, fontSize)
	defer impact.Close()

	// Get normalized forwarding
	drawer.Dot = fixed.P(0, 0)

	drawer.Face = impact

	bounds, advance := drawer.BoundString(s)
	height := bounds.Max.Y - bounds.Min.Y

	// Draw around p (center text in terms of p)
	// Calculate from perspective of baseline
	drawer.Dot = fixed.Point26_6{
		X: p.X - advance/2,
		Y: p.Y + height/2}

	for _, c := range s {
		// Render outline
		drawer.Src = image.Black

		// Save position before drawing (DrawString advances the Dot)
		begin := drawer.Dot

		outlineOffset := fixed.I(Outline)
		// Right
		drawer.Dot = begin.Add(fixed.Point26_6{
			X: +outlineOffset,
			Y: 0})
		drawer.DrawString(string(c))
		// Save position after being finished with the furthest character
		end := drawer.Dot

		// Bottom right
		drawer.Dot = begin.Add(fixed.Point26_6{
			X: 0,
			Y: -outlineOffset})
		drawer.DrawString(string(c))
		// Bottom
		drawer.Dot = begin.Add(fixed.Point26_6{
			X: -outlineOffset,
			Y: 0})
		drawer.DrawString(string(c))
		// Bottom left
		drawer.Dot = begin.Add(fixed.Point26_6{
			X: -outlineOffset,
			Y: 0})
		drawer.DrawString(string(c))
		// Left
		drawer.Dot = begin.Add(fixed.Point26_6{
			X: 0,
			Y: +outlineOffset})
		drawer.DrawString(string(c))
		// Top left
		drawer.Dot = begin.Add(fixed.Point26_6{
			X: 0,
			Y: +outlineOffset})
		drawer.DrawString(string(c))
		// Top
		drawer.Dot = begin.Add(fixed.Point26_6{
			X: +outlineOffset,
			Y: 0})
		drawer.DrawString(string(c))
		// Top right
		drawer.Dot = begin.Add(fixed.Point26_6{
			X: +outlineOffset,
			Y: 0})
		drawer.DrawString(string(c))

		// Restore original position (before outline)
		drawer.Dot = begin

		// White font color
		drawer.Src = image.White

		// Render inside
		drawer.DrawString(string(c))
		drawer.Dot = end
	}
}

func impacc(src string, dst string, top string, bottom string) {
	// Load font
	impactFont := initImpact()

	// Read image
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
	fontSize = float64(height) / FontProp

	// Read font and initialize faces

	// Set draw buffer
	drawImg := image.NewRGBA(bounds)
	draw.Draw(drawImg, bounds, img, image.Point{0, 0}, draw.Src)

	drawer := font.Drawer{
		Dst: drawImg}

	// Top text
	drawImpactStr(
		impactFont,
		&drawer,
		strings.ToUpper(top),
		fixed.P(width/2, int(fontSize/2)+Offset))

	// Bottom text
	drawImpactStr(
		impactFont,
		&drawer,
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
