package qrcode

import (
	"bytes"
	"encoding/xml"
	"image/png"
	"strings"
	"testing"
)

const sample = "https://example.com/AbCdEf12"

func TestPNG(t *testing.T) {
	data, err := PNG(sample)
	if err != nil {
		t.Fatal(err)
	}

	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("invalid PNG: %v", err)
	}
	if b := img.Bounds(); b.Dx() != PNGSize || b.Dy() != PNGSize {
		t.Fatalf("got %dx%d, want %dx%d", b.Dx(), b.Dy(), PNGSize, PNGSize)
	}

	// Corners belong to the quiet zone and must be white.
	if r, _, _, _ := img.At(0, 0).RGBA(); r == 0 {
		t.Error("top-left corner should be white")
	}
	// The center of the image is inside the code: it must contain black pixels.
	black := false
	for x := 0; x < PNGSize && !black; x++ {
		if r, _, _, _ := img.At(x, PNGSize/2).RGBA(); r == 0 {
			black = true
		}
	}
	if !black {
		t.Error("no black pixel on the middle row")
	}
}

func TestSVG(t *testing.T) {
	data, err := SVG(sample)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(data), `<svg xmlns="http://www.w3.org/2000/svg"`) {
		t.Fatalf("not an SVG document: %s", data)
	}
	var doc struct {
		XMLName xml.Name
	}
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid XML: %v", err)
	}
	if doc.XMLName.Local != "svg" {
		t.Fatalf("root element is %q", doc.XMLName.Local)
	}
}

func TestTooLong(t *testing.T) {
	if _, err := PNG(strings.Repeat("a", 5000)); err == nil {
		t.Error("expected an error for text exceeding QR capacity")
	}
}
