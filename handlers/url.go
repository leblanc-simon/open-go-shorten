package handlers

import (
	"encoding/json"
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
	json.NewDecoder(r.Body).Decode(&urlData)

	// Generate short URL
	shortURL := utils.GenerateShortURL(urlData.URL)

	// Store short URL, original URL, and expiration date in Redis
	err := utils.StoreURL(shortURL, urlData.URL, urlData.Expiration)
	if err != nil {
		http.Error(w, "Error storing URL", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"shortURL": shortURL})
}

func DeleteUrl(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("shortURL")

	err := utils.RemoveUrl(shortURL)
	if err != nil {
		http.Error(w, "URL can't be deleted", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"shortURL": shortURL})
}

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	shortURL := r.PathValue("shortURL")

	// Retrieve original URL from Redis
	originalURL, err := utils.GetOriginalURL(shortURL)
	if err != nil {
		http.Error(w, "URL not found or expired", http.StatusNotFound)
		return
	}

	// Extract IP address and user agent from request
	ip := clientIP(r)
	userAgent := r.UserAgent()

	// Create new Visit struct
	visit := models.Visit{
		Timestamp: time.Now(),
		IP:        ip,
		UserAgent: userAgent,
		ShortURL:  shortURL,
	}

	// Store visit event in Redis
	err = utils.StoreVisit(shortURL, visit)
	if err != nil {
		http.Error(w, "Error storing visit", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusMovedPermanently)
}
