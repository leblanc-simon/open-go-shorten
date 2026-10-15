package handlers

import (
	"log/slog"
	"net/http"

	"open-go-shorten.eu/utils"
)

func GetURLs(w http.ResponseWriter, r *http.Request) {
	// Query Redis for shortened URLs and their statistics
	urls, err := utils.GetUrls(r.Context())
	if err != nil {
		slog.Error("get URLs", "err", err)
		http.Error(w, "Error retrieving URLs", http.StatusInternalServerError)
		return
	}

	// Return list of shortened URLs and their statistics
	writeJSON(w, urls)
}
