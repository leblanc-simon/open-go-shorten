package utils

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"open-go-shorten.eu/config"
	"open-go-shorten.eu/models"
)

// maxCollisionAttempts bounds the search of a free short URL.
const maxCollisionAttempts = 10

var (
	ErrNotFound       = errors.New("short URL not found")
	ErrNoFreeShortURL = errors.New("no free short URL")
)

var rdb *redis.Client
var redisPrefix string

func InitRedis(c *config.Config) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     c.Database.Host + ":" + strconv.Itoa(c.Database.Port),
		Password: c.Database.Password,
		DB:       c.Database.Db,
	})
	redisPrefix = c.Database.Prefix
}

// PingRedis checks the connection to the Redis server.
func PingRedis(ctx context.Context) error {
	return rdb.Ping(ctx).Err()
}

func urlKey(shortURL string) string {
	return redisPrefix + shortURL
}

func visitsKey(shortURL string) string {
	return redisPrefix + "visits-" + shortURL
}

// StoreURL stores the original URL and returns its short URL. A nil expiration
// means that the short URL never expires. Shortening an already shortened URL
// returns the same short URL and updates its expiration.
func StoreURL(ctx context.Context, originalURL string, expiration *time.Time) (string, error) {
	var ttl time.Duration
	if expiration != nil {
		ttl = time.Until(*expiration)
		if ttl <= 0 {
			return "", errors.New("expiration date is in the past")
		}
	}

	for attempt := 0; attempt < maxCollisionAttempts; attempt++ {
		shortURL, err := GenerateShortURL(originalURL, attempt)
		if err != nil {
			return "", err
		}

		existing, err := rdb.Get(ctx, urlKey(shortURL)).Result()
		if err != nil && !errors.Is(err, redis.Nil) {
			return "", err
		}
		if err == nil && existing != originalURL {
			// Collision with another URL: never overwrite it.
			continue
		}

		return shortURL, rdb.Set(ctx, urlKey(shortURL), originalURL, ttl).Err()
	}

	return "", ErrNoFreeShortURL
}

func RemoveUrl(ctx context.Context, shortURL string) error {
	return rdb.Del(ctx, urlKey(shortURL), visitsKey(shortURL)).Err()
}

// GetOriginalURL returns the original URL of a short URL, or ErrNotFound.
func GetOriginalURL(ctx context.Context, shortURL string) (string, error) {
	originalURL, err := rdb.Get(ctx, urlKey(shortURL)).Result()
	if errors.Is(err, redis.Nil) {
		return "", ErrNotFound
	}
	return originalURL, err
}

func StoreVisit(ctx context.Context, shortURL string, visit models.Visit) error {
	jsonVisit, err := json.Marshal(visit)
	if err != nil {
		return err
	}
	return rdb.LPush(ctx, visitsKey(shortURL), jsonVisit).Err()
}

// GetUrls returns every short URL with its statistics.
func GetUrls(ctx context.Context) ([]models.History, error) {
	history := []models.History{}

	iter := rdb.Scan(ctx, 0, redisPrefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		shortURL := strings.TrimPrefix(key, redisPrefix)
		if strings.HasPrefix(shortURL, "visits-") {
			continue
		}

		originalURL, err := rdb.Get(ctx, key).Result()
		if err != nil {
			// Expired between SCAN and GET, or not a short URL.
			continue
		}

		ttl, err := rdb.TTL(ctx, key).Result()
		if err != nil {
			return nil, err
		}
		var expiration *time.Time
		if ttl > 0 {
			t := time.Now().Add(ttl)
			expiration = &t
		}

		stats, err := getStats(ctx, shortURL)
		if err != nil {
			return nil, err
		}

		history = append(history, models.History{
			URLData: models.URLData{URL: originalURL, Expiration: expiration},
			Stats:   stats,
			Shorten: shortURL,
		})
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}

	return history, nil
}

func getStats(ctx context.Context, shortURL string) (models.Stats, error) {
	visitsJSON, err := rdb.LRange(ctx, visitsKey(shortURL), 0, -1).Result()
	if err != nil {
		return models.Stats{}, err
	}

	uniqueVisitors := make(map[string]struct{})
	for _, jsonVisit := range visitsJSON {
		var visit models.Visit
		if err := json.Unmarshal([]byte(jsonVisit), &visit); err != nil {
			continue
		}
		uniqueVisitors[visit.IP] = struct{}{}
	}

	return models.Stats{
		VisitorsCount:       len(visitsJSON),
		UniqueVisitorsCount: len(uniqueVisitors),
	}, nil
}
