package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"story-backend/config"
	"story-backend/models"

	"github.com/labstack/echo/v4"
)

// Follow a user
func FollowUser(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "no token"})
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	userID, err := extractUserIDFromToken(tokenStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	targetIDParam := c.Param("id")
	targetID, err := strconv.Atoi(targetIDParam)
	if err != nil || targetID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	if uint(targetID) == userID {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "cannot follow yourself"})
	}

	follow := models.Follow{
		FollowerID: userID,
		FolloweeID: uint(targetID),
	}

	if err := config.DB.Create(&follow).Error; err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "already following or db error"})
	}

	return c.JSON(http.StatusOK, follow)
}

// Unfollow a user
func UnfollowUser(c echo.Context) error {
	authHeader := c.Request().Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "no token"})
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	userID, err := extractUserIDFromToken(tokenStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	targetIDParam := c.Param("id")
	targetID, err := strconv.Atoi(targetIDParam)
	if err != nil || targetID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	if err := config.DB.Delete(&models.Follow{}, "follower_id = ? AND followee_id = ?", userID, targetID).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "unfollowed"})
}

// Get list of users a user follows
func GetFollowing(c echo.Context) error {
	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil || userID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	var follows []models.Follow
	if err := config.DB.Where("follower_id = ?", userID).Find(&follows).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	return c.JSON(http.StatusOK, echo.Map{"following": follows})
}

// Get list of followers for a user
func GetFollowers(c echo.Context) error {
	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil || userID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	var follows []models.Follow
	if err := config.DB.Where("followee_id = ?", userID).Find(&follows).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	return c.JSON(http.StatusOK, echo.Map{"followers": follows})
}
