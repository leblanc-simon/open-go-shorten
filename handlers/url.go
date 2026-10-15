package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"open-go-shorten.eu/models"
	"open-go-shorten.eu/utils"
)

var clientIP = func(r *http.Request) string { return r.RemoteAddr }

// InitClientIP sets the function used to resolve the IP of the visitors.
func InitClientIP(f func(r *http.Request) string) {
	clientIP = f
}

func ShortenURL(w http.ResponseWriter, r *http.Request) {
	var urlData models.URLData
	if err := decodeJSON(w, r, &urlData); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if err := utils.ValidateURL(urlData.URL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if urlData.Expiration != nil && !urlData.Expiration.After(time.Now()) {
		http.Error(w, "Expiration date is in the past", http.StatusBadRequest)
		return
	}

	// Store short URL, original URL, and expiration date in Redis
	shortURL, err := utils.StoreURL(r.Context(), urlData.URL, urlData.Expiration)
	if err != nil {
		slog.Error("store URL", "url", urlData.URL, "err", err)
		http.Error(w, "Error storing URL", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"shortURL": shortURL})
}

func DeleteUrl(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("shortURL")
	if !utils.IsValidShortURL(shortURL) {
		http.NotFound(w, r)
		return
	}

	if err := utils.RemoveUrl(r.Context(), shortURL); err != nil {
		slog.Error("remove URL", "short_url", shortURL, "err", err)
		http.Error(w, "URL can't be deleted", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"shortURL": shortURL})
}

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("shortURL")
	if !utils.IsValidShortURL(shortURL) {
		http.NotFound(w, r)
		return
	}

	// Retrieve original URL from Redis
	originalURL, err := utils.GetOriginalURL(r.Context(), shortURL)
	if errors.Is(err, utils.ErrNotFound) {
		http.Error(w, "URL not found or expired", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.Error("get original URL", "short_url", shortURL, "err", err)
		http.Error(w, "Error retrieving URL", http.StatusInternalServerError)
		return
	}

	// Store visit event in Redis: a failure must not prevent the redirection
	visit := models.Visit{
		Timestamp: time.Now(),
		IP:        clientIP(r),
		UserAgent: r.UserAgent(),
		ShortURL:  shortURL,
	}
	if err := utils.StoreVisit(r.Context(), shortURL, visit); err != nil {
		slog.Error("store visit", "short_url", shortURL, "err", err)
	}

	// Temporary redirection: a permanent one would be cached by the browsers,
	// so the next visits would not be counted, and a deleted short URL would
	// still be followed.
	w.Header().Set("Cache-Control", "private, no-store")
	http.Redirect(w, r, originalURL, http.StatusFound)
}
