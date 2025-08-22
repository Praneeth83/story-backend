package utils

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"story-backend/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

//
// ------------------ PASSWORD HELPERS ------------------
//

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

func CheckPassword(hashed, plain string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain))
}

//
// ------------------ JWT HELPERS ------------------
//

func GenerateJWT(userID uint) (string, error) {
	return GenerateJWTWithExpiry(userID, 72*time.Hour)
}

func GenerateJWTWithExpiry(userID uint, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
		"iat":     time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.JWTSecret))
}

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
	// (Optional) Expiration check since using MapClaims
	if exp, ok := claims["exp"].(float64); ok && int64(exp) < time.Now().Unix() {
		return nil, errors.New("token expired")
	}
	return claims, nil
}

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

func SplitBearer(header string) string {
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

//
// ------------------ CONTEXT HELPERS ------------------
//

func GetUserID(c echo.Context) (uint, error) {
	uid, ok := c.Get("user_id").(uint)
	if !ok || uid == 0 {
		return 0, errors.New("unauthorized")
	}
	return uid, nil
}
