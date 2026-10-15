package middleware

import (
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"open-go-shorten.eu/config"
)

var jwtSecret []byte

func InitJwt(c *config.Config) {
	jwtSecret = []byte(c.Auth.JwtSecret)
}

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract JWT from Authorization header
		tokenString, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || tokenString == "" {
			http.Error(w, "Missing JWT", http.StatusUnauthorized)
			return
		}

		// Verify JWT: only accept the algorithm used to sign it, and require
		// an expiration date.
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
			return jwtSecret, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithExpirationRequired())

		if err != nil || !token.Valid {
			http.Error(w, "Invalid JWT", http.StatusUnauthorized)
			return
		}

		// Call next handler
		next.ServeHTTP(w, r)
	})
}
