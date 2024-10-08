package handlers

import (
	"encoding/json"
	"net/http"

	"open-go-shorten.eu/utils"
)

func GetURLs(w http.ResponseWriter, r *http.Request) {
	// Query Redis for shortened URLs and their statistics
	urls, err := utils.GetUrls()
	if err != nil {
		http.Error(w, "Error retrieving URLs", http.StatusInternalServerError)
	}

	// Return list of shortened URLs and their statistics
	json.NewEncoder(w).Encode(urls)
}
