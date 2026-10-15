package handlers

import (
	"crypto/subtle"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"open-go-shorten.eu/config"
	"open-go-shorten.eu/models"
)

var jwtSecret []byte
var username string
var password []byte

func InitAuth(c *config.Config) {
	jwtSecret = []byte(c.Auth.JwtSecret)
	username = c.Auth.Username
	password = []byte(c.Auth.Password)
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := decodeJSON(w, r, &user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Verify credentials (always check the password to not leak the username
	// through the response time)
	validUsername := subtle.ConstantTimeCompare([]byte(user.Username), []byte(username)) == 1
	validPassword := bcrypt.CompareHashAndPassword(password, []byte(user.Password)) == nil
	if !validUsername || !validPassword {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, "Error creating JWT", http.StatusInternalServerError)
		return
	}

	// Return JWT
	writeJSON(w, map[string]string{"token": tokenString})
}
