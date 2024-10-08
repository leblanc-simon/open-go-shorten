package models

import "time"

type URLData struct {
	URL        string    `json:"url"`
	Expiration time.Time `json:"expiration"`
}
