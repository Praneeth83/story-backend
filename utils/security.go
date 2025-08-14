package utils

import (
	"errors"
	"strconv"
	"time"

	"story-backend/config"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes a plain-text password using bcrypt
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword compares a bcrypt-hashed password with a plain-text password
func CheckPassword(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

// GenerateJWT generates a JWT for a given userID with 72-hour expiration
func GenerateJWT(userID uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

// GenerateJWTWithExpiry generates a JWT with a custom expiration duration
func GenerateJWTWithExpiry(userID uint, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
		"iat":     time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

// ExtractUserIDFromClaims safely extracts the user_id from MapClaims
func ExtractUserIDFromClaims(claims jwt.MapClaims) (uint, error) {
	uidVal, ok := claims["user_id"]
	if !ok {
		return 0, errInvalidUserID
	}

	switch v := uidVal.(type) {
	case float64:
		return uint(v), nil
	case int:
		return uint(v), nil
	case string:
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, errInvalidUserID
		}
		return uint(n), nil
	default:
		return 0, errInvalidUserID
	}
}

// Predefined error for invalid user_id in token
var errInvalidUserID = errors.New("invalid user_id in token")
