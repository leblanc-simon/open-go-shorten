package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"open-go-shorten.eu/config"
	"open-go-shorten.eu/qrcode"
	"open-go-shorten.eu/utils"
)

var publicURL string
var qrGenerator = &qrcode.Generator{}

func InitQRCode(c *config.Config) error {
	publicURL = strings.TrimRight(c.Server.PublicURL, "/")

	generator, err := qrcode.New(c.QRCode.Logo)
	if err != nil {
		return err
	}
	qrGenerator = generator
	return nil
}

// QRCode serves the QR code of a short link, as SVG or PNG (512x512).
// Add ?download=1 to get it as an attachment.
func QRCode(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("shortURL")
	format := r.PathValue("format")

	if !utils.IsValidShortURL(shortURL) {
		http.NotFound(w, r)
		return
	}
	if _, err := utils.GetOriginalURL(r.Context(), shortURL); err != nil {
		if errors.Is(err, utils.ErrNotFound) {
			http.Error(w, "URL not found or expired", http.StatusNotFound)
		} else {
			slog.Error("get original URL", "short_url", shortURL, "err", err)
			http.Error(w, "Error retrieving URL", http.StatusInternalServerError)
		}
		return
	}

	link := baseURL(r) + "/" + shortURL

	var (
		data        []byte
		err         error
		contentType string
	)
	switch format {
	case "svg":
		data, err = qrGenerator.SVG(link)
		contentType = "image/svg+xml"
	case "png":
		data, err = qrGenerator.PNG(link)
		contentType = "image/png"
	default:
		http.Error(w, "Unsupported format", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("generate QR code", "short_url", shortURL, "err", err)
		http.Error(w, "Error generating QR code", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, max-age=3600")
	if r.URL.Query().Has("download") {
		w.Header().Set("Content-Disposition", `attachment; filename="`+shortURL+`.`+format+`"`)
	}
	w.Write(data)
}

// baseURL returns the public base URL of the short links, without trailing
// slash: the configured one, or the one deduced from the request.
func baseURL(r *http.Request) string {
	if publicURL != "" {
		return publicURL
	}

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}
