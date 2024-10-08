package models

import "time"

type Visit struct {
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"userAgent"`
	ShortURL  string    `json:"shortURL"`
}
