package controllers

import (
	"net/http"
	"strconv"

	"story-backend/config"
	"story-backend/models"
	"story-backend/utils"

	"github.com/labstack/echo/v4"
)

// -------------------- Follow a user --------------------
// Follow a user (public -> auto follow, private -> request)
func FollowUser(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	identifier := c.Param("identifier")

	// Step 1: Find target user (by id or username)
	var target models.User
	if id, convErr := strconv.Atoi(identifier); convErr == nil {
		if err := config.DB.First(&target, id).Error; err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
		}
	} else {
		if err := config.DB.Where("username = ?", identifier).First(&target).Error; err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
		}
	}

	// Step 2: Prevent following yourself
	if target.ID == userID {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "cannot follow yourself"})
	}

	// Step 3: If account is private → create follow request
	if target.Type == "private" {
		req := models.FollowRequest{
			FollowerID: userID,
			FolloweeID: target.ID,
		}
		if err := config.DB.Create(&req).Error; err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "already requested"})
		}
		return c.JSON(http.StatusOK, echo.Map{"message": "follow request sent"})
	}

	// Step 4: If public → directly follow
	follow := models.Follow{
		FollowerID: userID,
		FolloweeID: target.ID,
	}
	if err := config.DB.Create(&follow).Error; err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "already following"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "followed"})
}

// -------------------- Unfollow a user --------------------
func UnfollowUser(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	identifier := c.Param("id")

	var target models.User
	if targetID, err := strconv.Atoi(identifier); err == nil {
		if err := config.DB.First(&target, targetID).Error; err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
		}
	} else {
		if err := config.DB.Where("username = ?", identifier).First(&target).Error; err != nil {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
		}
	}

	if err := config.DB.Delete(&models.Follow{}, "follower_id = ? AND followee_id = ?", userID, target.ID).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "unfollowed"})
}

// -------------------- Get following --------------------
func GetFollowing(c echo.Context) error {
	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil || userID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	type FollowingResponse struct {
		ID         uint   `json:"id"`
		Username   string `json:"username"`
		ProfilePic string `json:"profile_pic"`
	}

	var following []FollowingResponse
	if err := config.DB.
		Table("follows").
		Select("users.id, users.username, users.profile_pic").
		Joins("JOIN users ON follows.followee_id = users.id").
		Where("follows.follower_id = ?", userID).
		Scan(&following).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	return c.JSON(http.StatusOK, echo.Map{"following": following})
}

// -------------------- Get followers --------------------
func GetFollowers(c echo.Context) error {
	userIDParam := c.Param("id")
	userID, err := strconv.Atoi(userIDParam)
	if err != nil || userID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	type FollowerResponse struct {
		ID         uint   `json:"id"`
		Username   string `json:"username"`
		ProfilePic string `json:"profile_pic"`
	}

	var followers []FollowerResponse
	if err := config.DB.
		Table("follows").
		Select("users.id, users.username, users.profile_pic").
		Joins("JOIN users ON follows.follower_id = users.id").
		Where("follows.followee_id = ?", userID).
		Scan(&followers).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	return c.JSON(http.StatusOK, echo.Map{"followers": followers})
}

func GetFollowRequests(c echo.Context) error {
	userID, _ := utils.GetUserID(c)
	var requests []models.FollowRequest
	if err := config.DB.
		Where("followee_id = ?", userID).
		Find(&requests).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	return c.JSON(http.StatusOK, echo.Map{"requests": requests})
}

func AcceptFollowRequest(c echo.Context) error {
	followeeID, err := utils.GetUserID(c)
	if err != nil || followeeID == 0 {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "user not found"})
	}

	followerIDParam := c.Param("id")
	followerID, err := strconv.Atoi(followerIDParam)
	if err != nil || followerID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid follower id"})
	}

	var req models.FollowRequest
	if err := config.DB.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		First(&req).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "follow request not found"})
	}

	var existingFollow models.Follow
	if err := config.DB.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		First(&existingFollow).Error; err == nil {

		config.DB.Delete(&req)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "already following"})
	}

	follow := models.Follow{
		FollowerID: uint(followerID),
		FolloweeID: followeeID,
	}
	if err := config.DB.Create(&follow).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "could not create follow"})
	}

	config.DB.Delete(&req)

	return c.JSON(http.StatusOK, echo.Map{"message": "follow request accepted"})
}

func RejectFollowRequest(c echo.Context) error {
	// Get logged-in user (followee)
	followeeID, err := utils.GetUserID(c)
	if err != nil || followeeID == 0 {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "user not found"})
	}

	// Get follower ID from URL
	followerIDParam := c.Param("id")
	followerID, err := strconv.Atoi(followerIDParam)
	if err != nil || followerID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid follower id"})
	}

	// Find the follow request
	var req models.FollowRequest
	if err := config.DB.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).
		First(&req).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "follow request not found"})
	}

	// Delete the request (reject)
	if err := config.DB.Delete(&req).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "could not reject request"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "follow request rejected"})
}
