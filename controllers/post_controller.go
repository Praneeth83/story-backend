package controllers

import (
	"net/http"
	"strconv"
	"time"

	"story-backend/config"
	"story-backend/models"
	"story-backend/utils"

	"github.com/labstack/echo/v4"
)

// ---------- Feed: GET /posts/feed ----------
func GetPostsFeed(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	type row struct {
		PostID     uint      `json:"post_id"`
		UserID     uint      `json:"user_id"`
		Username   string    `json:"username"`
		ProfilePic string    `json:"profile_pic"`
		Caption    string    `json:"caption"`
		MediaURL   string    `json:"media_url"`
		MediaType  string    `json:"media_type"`
		CreatedAt  time.Time `json:"created_at"`
	}

	var rows []row
	err = config.DB.
		Table("posts AS p").
		Select(`
			p.id AS post_id,
			p.user_id,
			u.username,
			u.profile_pic,
			p.caption,
			p.media_url,
			p.media_type,
			p.created_at`).
		Joins("JOIN users AS u ON u.id = p.user_id").
		Joins("LEFT JOIN follows AS f ON f.followee_id = p.user_id AND f.follower_id = ?", userID).
		Where("(f.follower_id IS NOT NULL OR p.user_id = ?)", userID).
		Order("p.created_at DESC").
		Scan(&rows).Error

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to build feed"})
	}

	return c.JSON(http.StatusOK, rows)
}

// ---------- User Posts: GET /posts/user/:id ----------
func GetUserPosts(c echo.Context) error {
	targetIDParam := c.Param("id")
	targetID, err := strconv.Atoi(targetIDParam)
	if err != nil || targetID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	var rows []models.Post
	if err := config.DB.Where("user_id = ?", targetID).Order("created_at DESC").Find(&rows).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch posts"})
	}

	return c.JSON(http.StatusOK, rows)
}

// ---------- Add Post: POST /posts/add ----------
type addPostReq struct {
	Caption   string `json:"caption"`
	MediaURL  string `json:"media_url"`
	MediaType string `json:"media_type"`
}

func AddPost(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	var req addPostReq
	if err := c.Bind(&req); err != nil || req.MediaURL == "" || req.MediaType == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}

	post := models.Post{
		UserID:    userID,
		Caption:   req.Caption,
		MediaURL:  req.MediaURL,
		MediaType: req.MediaType,
		CreatedAt: time.Now(),
	}

	if err := config.DB.Create(&post).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create post"})
	}

	return c.JSON(http.StatusCreated, post)
}

// ---------- Delete Post: DELETE /posts/:id ----------
func DeletePost(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	id := c.Param("id")
	var post models.Post
	if err := config.DB.First(&post, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "post not found or not yours"})
	}

	if err := config.DB.Delete(&post).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to delete post"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "deleted"})
}
