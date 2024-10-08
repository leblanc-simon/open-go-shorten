package utils

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"open-go-shorten.eu/config"
	"open-go-shorten.eu/models"
)

var ctx = context.Background()
var rdb *redis.Client
var redisPrefix string

func InitRedis(c *config.Config) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     c.Database.Host + ":" + strconv.Itoa(c.Database.Port),
		Password: c.Database.Password, // no password set
		DB:       c.Database.Database, // use default DB
	})
	redisPrefix = c.Database.Prefix
}

func StoreURL(shortURL, originalURL string, expiration time.Time) error {
	if expiration.IsZero() {
		return rdb.Set(ctx, redisPrefix+shortURL, originalURL, 0).Err()
	} else {
		return rdb.Set(ctx, redisPrefix+shortURL, originalURL, time.Until(expiration)).Err()
	}
}

func RemoveUrl(shortURL string) error {
	rdb.Del(ctx, redisPrefix+"visits-"+shortURL).Err()
	return rdb.Del(ctx, redisPrefix+shortURL).Err()
}

func GetOriginalURL(shortURL string) (string, error) {
	return rdb.Get(ctx, redisPrefix+shortURL).Result()
}

func StoreVisit(shortURL string, visit models.Visit) error {
	jsonVisit, _ := json.Marshal(visit)
	return rdb.LPush(ctx, redisPrefix+"visits-"+shortURL, jsonVisit).Err()
}

func GetUrls() ([]models.History, error) {
	var history []models.History

	urlsJSON, err := rdb.Keys(ctx, "ogs-*").Result()
	if err != nil {
		return nil, err
	}

	for _, shortURL := range urlsJSON {
		regexForPrefix := regexp.MustCompile(`\bogs-\b`)
		realShortURL := regexForPrefix.ReplaceAllLiteralString(shortURL, "")

		originalURL, err := rdb.Get(ctx, shortURL).Result()
		if err != nil {
			continue
		}

		// Query Redis for visit events for the given shortened URL
		visitsJSON, err := rdb.LRange(ctx, redisPrefix+"visits-"+realShortURL, 0, -1).Result()
		if err != nil {
			continue
		}

		// Parse visit events and filter for the given shortened URL
		var visits []models.Visit
		for _, jsonVisit := range visitsJSON {
			var visit models.Visit
			json.Unmarshal([]byte(jsonVisit), &visit)
			visits = append(visits, visit)
		}

		// Calculate statistics
		totalVisits := len(visits)
		uniqueVisitors := make(map[string]models.Visit)
		for _, visit := range visits {
			uniqueVisitors[visit.IP] = visit
		}
		uniqueVisitorsCount := len(uniqueVisitors)

		// Create URLData struct
		var urlData models.URLData
		if rdb.TTL(ctx, redisPrefix+realShortURL).Val() == -1 {
			urlData = models.URLData{
				URL:        originalURL,
				Expiration: time.Unix(0, 1),
			}
		} else {
			urlData = models.URLData{
				URL:        originalURL,
				Expiration: time.Now().Add(rdb.TTL(ctx, redisPrefix+realShortURL).Val()),
			}
		}

		// Create Stats struct
		stats := models.Stats{
			VisitorsCount:       totalVisits,
			UniqueVisitorsCount: uniqueVisitorsCount,
			Visitors:            visits,
		}

		// Add URLData and Stats to list of URLs
		history = append(history, models.History{
			URLData: urlData,
			Stats:   stats,
			Shorten: realShortURL,
		})
	}

	return history, nil
}
