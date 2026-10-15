package qrcode

import (
	"bytes"
	"encoding/xml"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"path/filepath"
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

// logoPNG returns a red square PNG image.
func logoPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 40, 40))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.RGBA{R: 255, A: 255}), image.Point{}, draw.Src)
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}

func TestLogoPNG(t *testing.T) {
	g, err := NewWithLogo(logoPNG(t))
	if err != nil {
		t.Fatal(err)
	}
	if !g.HasLogo() {
		t.Fatal("HasLogo() = false")
	}

	data, err := g.PNG(sample)
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
	r, g2, b, _ := img.At(PNGSize/2, PNGSize/2).RGBA()
	if r < 0xf000 || g2 > 0x1000 || b > 0x1000 {
		t.Errorf("center pixel is not the red logo: %v %v %v", r, g2, b)
	}
}

func TestLogoSVG(t *testing.T) {
	g, err := NewWithLogo(logoPNG(t))
	if err != nil {
		t.Fatal(err)
	}

	data, err := g.SVG(sample)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `href="data:image/png;base64,`) {
		t.Error("logo not embedded in the SVG")
	}
	var doc struct{ XMLName xml.Name }
	if err := xml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid XML: %v", err)
	}
}

func TestLogoErrors(t *testing.T) {
	if _, err := New(filepath.Join(t.TempDir(), "missing.png")); err == nil {
		t.Error("expected an error for a missing file")
	}
	if _, err := NewWithLogo([]byte("<svg></svg>")); err == nil {
		t.Error("expected an error for an unsupported format")
	}
	if _, err := NewWithLogo([]byte("\x89PNG\r\n\x1a\ncorrupted")); err == nil {
		t.Error("expected an error for a corrupted image")
	}

	g, err := New("")
	if err != nil || g.HasLogo() {
		t.Errorf("New(\"\") = %v, %v: want a plain generator", g, err)
	}
}
