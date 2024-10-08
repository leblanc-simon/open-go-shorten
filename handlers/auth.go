package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"open-go-shorten.eu/config"
	"open-go-shorten.eu/models"
)

var jwtSecret string
var username string
var password string

func InitAuth(c *config.Config) {
	jwtSecret = c.Auth.JwtSecret
	username = c.Auth.Username
	password = c.Auth.Password
}

func Login(w http.ResponseWriter, r *http.Request) {
	var user models.User
	json.NewDecoder(r.Body).Decode(&user)

	// Verify credentials
	if user.Username != username || nil != bcrypt.CompareHashAndPassword([]byte(password), []byte(user.Password)) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		http.Error(w, "Error creating JWT", http.StatusInternalServerError)
		return
	}

	// Return JWT
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
