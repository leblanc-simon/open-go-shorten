package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// maxBodySize limits the size of the JSON request bodies.
const maxBodySize = 64 << 10

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	return json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(v)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("write JSON response", "err", err)
	}
}
