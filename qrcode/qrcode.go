// Package qrcode renders QR codes as SVG or PNG images.
package qrcode

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"

	"rsc.io/qr"
)

// PNGSize is the width and height, in pixels, of the generated PNG images.
const PNGSize = 512

// quietZone is the number of blank modules around the code, as required by
// the QR code specification.
const quietZone = 4

func encode(text string) (*qr.Code, error) {
	code, err := qr.Encode(text, qr.M)
	if err != nil {
		return nil, fmt.Errorf("qrcode: %w", err)
	}
	return code, nil
}

// SVG returns the QR code of text as a scalable SVG document.
func SVG(text string) ([]byte, error) {
	code, err := encode(text)
	if err != nil {
		return nil, err
	}

	total := code.Size + 2*quietZone
	var b bytes.Buffer
	fmt.Fprintf(&b, `<?xml version="1.0" encoding="UTF-8"?>`+"\n")
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %d %d" width="%d" height="%d" shape-rendering="crispEdges">`, total, total, PNGSize, PNGSize)
	fmt.Fprintf(&b, `<rect width="100%%" height="100%%" fill="#fff"/>`)
	b.WriteString(`<path fill="#000" d="`)
	for y := 0; y < code.Size; y++ {
		for x := 0; x < code.Size; x++ {
			if !code.Black(x, y) {
				continue
			}
			// Merge horizontal runs of black modules into a single rectangle.
			start := x
			for x+1 < code.Size && code.Black(x+1, y) {
				x++
			}
			fmt.Fprintf(&b, "M%d %dh%dv1h-%dz", start+quietZone, y+quietZone, x-start+1, x-start+1)
		}
	}
	b.WriteString(`"/></svg>`)

	return b.Bytes(), nil
}

// PNG returns the QR code of text as a PNGSize x PNGSize PNG image.
//
// Modules are drawn with an integer number of pixels to stay sharp, and the
// code is centered in the image (the remaining space extends the quiet zone).
func PNG(text string) ([]byte, error) {
	code, err := encode(text)
	if err != nil {
		return nil, err
	}

	scale := PNGSize / (code.Size + 2*quietZone)
	if scale < 1 {
		return nil, fmt.Errorf("qrcode: text too long to fit in %dx%d pixels", PNGSize, PNGSize)
	}
	offset := (PNGSize - code.Size*scale) / 2

	img := image.NewPaletted(image.Rect(0, 0, PNGSize, PNGSize), color.Palette{color.White, color.Black})
	for y := 0; y < code.Size; y++ {
		for x := 0; x < code.Size; x++ {
			if !code.Black(x, y) {
				continue
			}
			for dy := 0; dy < scale; dy++ {
				row := img.Pix[(offset+y*scale+dy)*img.Stride:]
				for dx := 0; dx < scale; dx++ {
					row[offset+x*scale+dx] = 1
				}
			}
		}
	}

	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return nil, fmt.Errorf("qrcode: %w", err)
	}
	return b.Bytes(), nil
}
