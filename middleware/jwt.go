package middleware

import (
	"net/http"
	"story-backend/utils"

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
			// 2. Extract token from Bearer
			tokenString := utils.SplitBearer(authHeader)
			if tokenString == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token format"})
			}

			// 3. Parse token
			claims, err := utils.ParseToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
			}

			// 4. Extract user_id
			userID, err := utils.ExtractUserID(claims)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
			}

			// 5. Store in context
			c.Set("user_id", userID)
			return next(c)
		}
	}
}
