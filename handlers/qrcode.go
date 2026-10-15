package handlers

import (
	"net/http"
	"strings"

	"open-go-shorten.eu/config"
	"open-go-shorten.eu/qrcode"
	"open-go-shorten.eu/utils"
)

var publicURL string

func InitQRCode(c *config.Config) {
	publicURL = strings.TrimRight(c.Server.PublicURL, "/")
}

// QRCode serves the QR code of a short link, as SVG or PNG (512x512).
// Add ?download=1 to get it as an attachment.
func QRCode(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("shortURL")
	format := r.PathValue("format")

	if _, err := utils.GetOriginalURL(shortURL); err != nil {
		http.Error(w, "URL not found or expired", http.StatusNotFound)
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
		data, err = qrcode.SVG(link)
		contentType = "image/svg+xml"
	case "png":
		data, err = qrcode.PNG(link)
		contentType = "image/png"
	default:
		http.Error(w, "Unsupported format", http.StatusNotFound)
		return
	}
	if err != nil {
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
