// Package qrcode renders QR codes as SVG or PNG images, optionally with a
// logo in their center.
package qrcode

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // logo format
	"image/png"
	"net/http"
	"os"

	"golang.org/x/image/draw"
	"rsc.io/qr"
)

// PNGSize is the width and height, in pixels, of the generated PNG images.
const PNGSize = 512

// quietZone is the number of blank modules around the code, as required by
// the QR code specification.
const quietZone = 4

// logoRatio is the width of the logo area relative to the width of the code.
// With the highest error correction level (30% of the code can be restored),
// a logo area of 30% of the width (about 9% of the surface) keeps the code
// readable.
const logoRatio = 0.30

// Generator renders QR codes. The zero value renders plain QR codes.
type Generator struct {
	logo        image.Image
	logoDataURI string
}

// New returns a Generator drawing the image of logoPath (PNG or JPEG) in the
// center of the QR codes. An empty logoPath gives plain QR codes.
func New(logoPath string) (*Generator, error) {
	if logoPath == "" {
		return &Generator{}, nil
	}

	data, err := os.ReadFile(logoPath)
	if err != nil {
		return nil, fmt.Errorf("qrcode: read logo: %w", err)
	}
	return NewWithLogo(data)
}

// NewWithLogo returns a Generator drawing the given image (PNG or JPEG) in
// the center of the QR codes.
func NewWithLogo(data []byte) (*Generator, error) {
	contentType := http.DetectContentType(data)
	if contentType != "image/png" && contentType != "image/jpeg" {
		return nil, fmt.Errorf("qrcode: logo must be a PNG or JPEG image, got %s", contentType)
	}

	logo, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("qrcode: decode logo: %w", err)
	}
	if b := logo.Bounds(); b.Dx() == 0 || b.Dy() == 0 {
		return nil, fmt.Errorf("qrcode: empty logo")
	}

	return &Generator{
		logo:        logo,
		logoDataURI: "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data),
	}, nil
}

// HasLogo reports whether the QR codes have a logo.
func (g *Generator) HasLogo() bool {
	return g.logo != nil
}

// layout is a QR code with the area reserved for the logo, in modules.
type layout struct {
	code *qr.Code
	// hole is the area of the logo, including a blank margin: no module is
	// drawn there. Empty without logo.
	hole image.Rectangle
	// logo is the area where the logo is drawn, inside hole.
	logo image.Rectangle
}

func (l layout) black(x, y int) bool {
	return l.code.Black(x, y) && !image.Pt(x, y).In(l.hole)
}

func (g *Generator) encode(text string) (layout, error) {
	level := qr.M
	if g.HasLogo() {
		level = qr.H
	}
	code, err := qr.Encode(text, level)
	if err != nil {
		return layout{}, fmt.Errorf("qrcode: %w", err)
	}

	l := layout{code: code}
	if !g.HasLogo() {
		return l, nil
	}

	// The code has an odd number of modules: an odd hole stays centered.
	side := int(float64(code.Size)*logoRatio) | 1
	start := (code.Size - side) / 2
	l.hole = image.Rect(start, start, start+side, start+side)
	l.logo = l.hole.Inset(1)
	return l, nil
}

// fit returns the largest rectangle with the aspect ratio of the logo,
// centered in area.
func (g *Generator) fit(area image.Rectangle) image.Rectangle {
	b := g.logo.Bounds()
	w, h := area.Dx(), area.Dy()
	if b.Dx()*h > b.Dy()*w {
		h = w * b.Dy() / b.Dx()
	} else {
		w = h * b.Dx() / b.Dy()
	}
	min := area.Min.Add(image.Pt((area.Dx()-w)/2, (area.Dy()-h)/2))
	return image.Rectangle{Min: min, Max: min.Add(image.Pt(w, h))}
}

// SVG returns the QR code of text as a scalable SVG document.
func (g *Generator) SVG(text string) ([]byte, error) {
	l, err := g.encode(text)
	if err != nil {
		return nil, err
	}

	size := l.code.Size
	total := size + 2*quietZone
	var b bytes.Buffer
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" shape-rendering="crispEdges">`, total, total, PNGSize, PNGSize)
	fmt.Fprintf(&b, `<rect width="100%%" height="100%%" fill="#fff"/>`)
	b.WriteString(`<path fill="#000" d="`)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !l.black(x, y) {
				continue
			}
			// Merge horizontal runs of black modules into a single rectangle.
			start := x
			for x+1 < size && l.black(x+1, y) {
				x++
			}
			fmt.Fprintf(&b, "M%d %dh%dv1h-%dz", start+quietZone, y+quietZone, x-start+1, x-start+1)
		}
	}
	b.WriteString(`"/>`)

	if g.HasLogo() {
		r := l.logo.Add(image.Pt(quietZone, quietZone))
		fmt.Fprintf(&b, `<image x="%d" y="%d" width="%d" height="%d" preserveAspectRatio="xMidYMid meet" href="%s"/>`,
			r.Min.X, r.Min.Y, r.Dx(), r.Dy(), g.logoDataURI)
	}
	b.WriteString(`</svg>`)

	return b.Bytes(), nil
}

// PNG returns the QR code of text as a PNGSize x PNGSize PNG image.
//
// Modules are drawn with an integer number of pixels to stay sharp, and the
// code is centered in the image (the remaining space extends the quiet zone).
func (g *Generator) PNG(text string) ([]byte, error) {
	l, err := g.encode(text)
	if err != nil {
		return nil, err
	}

	size := l.code.Size
	scale := PNGSize / (size + 2*quietZone)
	if scale < 1 {
		return nil, fmt.Errorf("qrcode: text too long to fit in %dx%d pixels", PNGSize, PNGSize)
	}
	offset := (PNGSize - size*scale) / 2
	bounds := image.Rect(0, 0, PNGSize, PNGSize)

	var img draw.Image
	if g.HasLogo() {
		rgba := image.NewRGBA(bounds)
		draw.Draw(rgba, bounds, image.White, image.Point{}, draw.Src)
		img = rgba
	} else {
		img = image.NewPaletted(bounds, color.Palette{color.White, color.Black})
	}

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if !l.black(x, y) {
				continue
			}
			module := image.Rect(x*scale, y*scale, (x+1)*scale, (y+1)*scale).Add(image.Pt(offset, offset))
			draw.Draw(img, module, image.Black, image.Point{}, draw.Src)
		}
	}

	if g.HasLogo() {
		area := image.Rectangle{Min: l.logo.Min.Mul(scale), Max: l.logo.Max.Mul(scale)}.Add(image.Pt(offset, offset))
		draw.CatmullRom.Scale(img, g.fit(area), g.logo, g.logo.Bounds(), draw.Over, nil)
	}

	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return nil, fmt.Errorf("qrcode: %w", err)
	}
	return b.Bytes(), nil
}

// SVG returns the plain QR code of text as a scalable SVG document.
func SVG(text string) ([]byte, error) {
	return (&Generator{}).SVG(text)
}

// PNG returns the plain QR code of text as a PNGSize x PNGSize PNG image.
func PNG(text string) ([]byte, error) {
	return (&Generator{}).PNG(text)
}
