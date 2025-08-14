package middleware

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"story-backend/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// 1. Get Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "missing token"})
			}

			// 2. Check Bearer format
			parts := splitBearer(authHeader)
			if parts == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token format"})
			}
			tokenString := parts

			// 3. Parse token using config.JWTSecret
			token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
				return []byte(config.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid or expired token"})
			}

			// 4. Extract claims safely
			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token claims"})
			}

			// 5. Check expiration manually
			if exp, ok := claims["exp"].(float64); ok {
				if int64(exp) < time.Now().Unix() {
					return c.JSON(http.StatusUnauthorized, echo.Map{"error": "token expired"})
				}
			} else {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token expiration"})
			}

			// 6. Extract user_id safely
			userID, err := extractUserIDFromClaims(claims)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
			}

			c.Set("user_id", userID)
			return next(c)
		}
	}
}

// Split Bearer token safely
func splitBearer(header string) string {
	parts := strings.Split(header, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}
	return parts[1]
}

// Extract user_id safely from claims
func extractUserIDFromClaims(claims jwt.MapClaims) (uint, error) {
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
