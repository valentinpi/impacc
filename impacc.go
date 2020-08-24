package main

import (
	"flag"
	"fmt"
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
	// OutlineProp Thickness of text outline
	OutlineProp = 0.10
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
	var options truetype.Options
	options.Size = size

	face := truetype.NewFace(font, &options)
	return face
}

// Does NOT preserve the drawer given
// Only requires the Dst field to be initialized with the output buffer
func drawImpactStr(impactFont *truetype.Font, drawer *font.Drawer, s string, p fixed.Point26_6) {
	outlineOffset := fontSize * OutlineProp
	impact := initFace(impactFont, fontSize-outlineOffset)
	impactOutline := initFace(impactFont, fontSize)
	defer impact.Close()
	defer impactOutline.Close()

	// Get normalized forwarding
	drawer.Dot = fixed.P(0, 0)

	// The outline is bigger
	drawer.Face = impactOutline

	bounds, advance := drawer.BoundString(s)
	height := bounds.Max.Y - bounds.Min.Y

	// Draw around p (center text in terms of p)
	// Calculate from perspective of baseline
	drawer.Dot = fixed.Point26_6{
		X: p.X - advance/2,
		Y: p.Y + height/2}

	for _, c := range s {
		// Render outline
		drawer.Face = impactOutline
		drawer.Src = image.Black

		_, advanceOut := drawer.BoundString(string(c))

		// Save position before drawing (DrawString advances the Dot)
		begin := drawer.Dot

		drawer.DrawString(string(c))

		// Save position after being finished for the next character
		end := drawer.Dot

		// Render inside
		drawer.Face = impact
		drawer.Src = image.White

		// Restore original position (before outline)
		drawer.Dot = begin

		_, advanceIn := drawer.BoundString(string(c))

		// Add offset from outline
		offsetIn := fixed.Point26_6{
			X: (advanceOut - advanceIn).Mul(fixed.Int26_6(1 << 5)),
			Y: -fixed.I(int(outlineOffset / 2))}
		drawer.Dot = drawer.Dot.Add(offsetIn)
		fmt.Println(offsetIn)

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
