package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"open-go-shorten.eu/models"
	"open-go-shorten.eu/utils"
)

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
	vars := mux.Vars(r)
	shortURL := vars["shortURL"]

	err := utils.RemoveUrl(shortURL)
	if err != nil {
		http.Error(w, "URL can't be deleted", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"shortURL": shortURL})
}

func RedirectURL(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortURL := vars["shortURL"]

	// Retrieve original URL from Redis
	originalURL, err := utils.GetOriginalURL(shortURL)
	if err != nil {
		http.Error(w, "URL not found or expired", http.StatusNotFound)
		return
	}

	// Extract IP address and user agent from request
	ip := utils.GetIP(r)
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
