package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"story-backend/config"
	"story-backend/models"
	"story-backend/utils"

	"github.com/labstack/echo/v4"
)

// ---------- Feed: GET /stories/feed ----------
// Returns stories for the logged-in user
// ---------- Feed: GET /stories/feed ----------
// Returns stories from users the logged-in user follows
func GetStoriesFeed(c echo.Context) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	type row struct {
		UserID     uint
		Username   string
		ProfilePic string
		StoryID    uint
		MediaURL   string
		MediaType  string
		ViewedID   *uint
	}

	var rows []row
	err = config.DB.
		Table("stories AS s").
		Select(`
			u.id AS user_id,
			u.username,
			u.profile_pic,
			s.id AS story_id,
			s.media_url,
			s.media_type,
			sv.id AS viewed_id`).
		// Join users table
		Joins("JOIN users AS u ON u.id = s.user_id").
		// Left join follows so we can filter by either follow relationship OR own stories
		Joins("LEFT JOIN follows AS f ON f.followee_id = s.user_id AND f.follower_id = ?", userID).
		// Keep only stories that belong to someone I follow OR myself
		Where("(f.follower_id IS NOT NULL OR s.user_id = ?)", userID).
		Where("s.expires_at > ?", time.Now()).
		Joins("LEFT JOIN story_views AS sv ON sv.story_id = s.id AND sv.viewer_id = ?", userID).
		Order("u.id ASC, s.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to build feed"})
	}

	type storyItem struct {
		ID        uint   `json:"id"`
		MediaURL  string `json:"media_url"`
		MediaType string `json:"media_type"`
		Seen      bool   `json:"seen"`
	}
	type userBlock struct {
		UserID     uint        `json:"user_id"`
		Username   string      `json:"username"`
		ProfilePic string      `json:"profile_pic"`
		Stories    []storyItem `json:"stories"`
		AllSeen    bool        `json:"all_seen"`
	}

	feedMap := make(map[uint]*userBlock)
	for _, r := range rows {
		block, ok := feedMap[r.UserID]
		if !ok {
			block = &userBlock{
				UserID:     r.UserID,
				Username:   r.Username,
				ProfilePic: r.ProfilePic,
				Stories:    []storyItem{},
				AllSeen:    true,
			}
			feedMap[r.UserID] = block
		}
		seen := r.ViewedID != nil
		if !seen {
			block.AllSeen = false
		}
		block.Stories = append(block.Stories, storyItem{
			ID:        r.StoryID,
			MediaURL:  r.MediaURL,
			MediaType: r.MediaType,
			Seen:      seen,
		})
	}

	out := make([]userBlock, 0, len(feedMap))
	var lastUserID uint
	for _, r := range rows {
		if r.UserID != lastUserID {
			out = append(out, *feedMap[r.UserID])
			lastUserID = r.UserID
		}
	}
	return c.JSON(http.StatusOK, out)
}

// GET /stories/user/:id
func GetUserStories(c echo.Context) error {
	currentUserID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	targetIDParam := c.Param("id")
	targetID, err := strconv.Atoi(targetIDParam)
	if err != nil || targetID <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	// ✅ Check if the user exists
	var exists bool
	if err := config.DB.Model(&models.User{}).
		Select("count(*) > 0").
		Where("id = ?", targetID).
		Find(&exists).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "database error"})
	}
	if !exists {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
	}

	type row struct {
		StoryID    uint
		MediaURL   string
		MediaType  string
		ViewedID   *uint
		UserID     uint
		Username   string
		ProfilePic string
	}

	var rows []row
	err = config.DB.
		Table("stories AS s").
		Select(`
			s.id AS story_id,
			s.media_url,
			s.media_type,
			sv.id AS viewed_id,
			u.id AS user_id,
			u.username,
			u.profile_pic`).
		Joins("JOIN users AS u ON u.id = s.user_id").
		Joins("LEFT JOIN story_views AS sv ON sv.story_id = s.id AND sv.viewer_id = ?", currentUserID).
		Where("s.user_id = ?", targetID).
		Where("s.expires_at > ?", time.Now()).
		Order("s.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to fetch stories"})
	}

	type storyItem struct {
		ID        uint   `json:"id"`
		MediaURL  string `json:"media_url"`
		MediaType string `json:"media_type"`
		Seen      bool   `json:"seen"`
	}

	out := struct {
		UserID     uint        `json:"user_id"`
		Username   string      `json:"username"`
		ProfilePic string      `json:"profile_pic"`
		Stories    []storyItem `json:"stories"`
	}{}

	if len(rows) > 0 {
		out.UserID = rows[0].UserID
		out.Username = rows[0].Username
		out.ProfilePic = rows[0].ProfilePic
	}

	for _, r := range rows {
		out.Stories = append(out.Stories, storyItem{
			ID:        r.StoryID,
			MediaURL:  r.MediaURL,
			MediaType: r.MediaType,
			Seen:      r.ViewedID != nil,
		})
	}

	return c.JSON(http.StatusOK, out)
}

// ---------- Add Story: POST /stories/add ----------
type addStoryReq struct {
	MediaURL   string `json:"media_url"`
	MediaType  string `json:"media_type"`
	TTLMinutes int    `json:"ttl_minutes,omitempty"`
}

func AddStory(c echo.Context) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	var req addStoryReq
	if err := c.Bind(&req); err != nil || req.MediaURL == "" || req.MediaType == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}

	ttl := 24 * time.Hour
	if req.TTLMinutes > 0 {
		ttl = time.Duration(req.TTLMinutes) * time.Minute
	}

	st := models.Story{
		UserID:    userID,
		MediaURL:  req.MediaURL,
		MediaType: req.MediaType,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}
	if err := config.DB.Create(&st).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to create story"})
	}

	return c.JSON(http.StatusCreated, st)
}

// ---------- Delete Story: DELETE /stories/:id ----------
func DeleteStory(c echo.Context) error {
	userID, err := getUserIDFromToken(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	id := c.Param("id")
	var story models.Story
	if err := config.DB.First(&story, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "story not found or not yours"})
	}

	if err := config.DB.Delete(&story).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to delete story"})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "deleted"})
}

// ---------- Helper: get user_id from JWT ----------
func getUserIDFromToken(c echo.Context) (uint, error) {
	authHeader := c.Request().Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return 0, echo.ErrUnauthorized
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := utils.ParseToken(tokenStr)
	if err != nil {
		return 0, echo.ErrUnauthorized
	}

	uidVal, ok := claims["user_id"]
	if !ok {
		return 0, echo.ErrUnauthorized
	}

	switch v := uidVal.(type) {
	case float64:
		return uint(v), nil
	case int:
		return uint(v), nil
	case string:
		n, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return 0, echo.ErrUnauthorized
		}
		return uint(n), nil
	default:
		return 0, echo.ErrUnauthorized
	}
}
