package utils

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"strconv"

	"github.com/itchyny/base58-go"
)

// shortURLLength is the length of the generated short URLs.
const shortURLLength = 8

var shortURLPattern = regexp.MustCompile(`^[1-9A-HJ-NP-Za-km-z]{1,32}$`)

// IsValidShortURL reports whether s can be a short URL (base58 characters only).
func IsValidShortURL(s string) bool {
	return shortURLPattern.MatchString(s)
}

// GenerateShortURL returns the short URL of the original URL. The result is
// deterministic; attempt > 0 gives alternative values, used on collision.
func GenerateShortURL(originalURL string, attempt int) (string, error) {
	input := originalURL
	if attempt > 0 {
		input += "#" + strconv.Itoa(attempt)
	}

	hash := sha256.Sum256([]byte(input))
	number := new(big.Int).SetBytes(hash[:]).Uint64()
	encoded, err := base58.BitcoinEncoding.Encode([]byte(strconv.FormatUint(number, 10)))
	if err != nil {
		return "", fmt.Errorf("encode short URL: %w", err)
	}
	return string(encoded[:shortURLLength]), nil
}

// ValidateURL checks that raw is an absolute http(s) URL.
func ValidateURL(raw string) error {
	if len(raw) > 4096 {
		return fmt.Errorf("URL too long")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("only http and https URLs are allowed")
	}
	if u.Host == "" {
		return fmt.Errorf("URL without host")
	}
	return nil
}
