package models

import "time"

type URLData struct {
	URL string `json:"url"`
	// Expiration is nil when the short URL never expires.
	Expiration *time.Time `json:"expiration"`
}
