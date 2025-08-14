package utils

import (
	"errors"
	"story-backend/config"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func ParseToken(tokenStr string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(config.JWTSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	// Manual expiration check
	if exp, ok := claims["exp"].(float64); ok {
		if int64(exp) < time.Now().Unix() {
			return nil, errors.New("token expired")
		}
	} else {
		return nil, errors.New("invalid token expiration")
	}

	return claims, nil
}

// Helper to safely extract user_id from claims
func ExtractUserID(claims map[string]interface{}) (uint, error) {
	uidVal, ok := claims["user_id"]
	if !ok {
		return 0, errors.New("user_id not found in token")
	}

	switch v := uidVal.(type) {
	case float64:
		return uint(v), nil
	case int:
		return uint(v), nil
	case string:
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, errors.New("invalid user_id in token")
		}
		return uint(n), nil
	default:
		return 0, errors.New("invalid user_id type in token")
	}
}
