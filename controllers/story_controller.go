package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"story-backend/config"
	"story-backend/models"
	"story-backend/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// ---------- Feed: GET /stories/feed ----------
// Returns stories from users the logged-in user follows
func GetStoriesFeed(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	// Struct for scanning raw query results
	type row struct {
		UserID     uint
		Username   string
		ProfilePic string
		StoryID    uint
		MediaURL   string
		MediaType  string
		CreatedAt  time.Time
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
			s.created_at,
			sv.id AS viewed_id`).
		// Join users table
		Joins("JOIN users AS u ON u.id = s.user_id").
		// Left join follows so we can filter by either follow relationship OR own stories
		Joins("LEFT JOIN follows AS f ON f.followee_id = s.user_id AND f.follower_id = ?", userID).
		// Keep only stories that belong to someone I follow OR myself
		Where("(f.follower_id IS NOT NULL OR s.user_id = ?)", userID).
		Where("s.expires_at > ?", time.Now()).
		// Check if current user has viewed story
		Joins("LEFT JOIN story_views AS sv ON sv.story_id = s.id AND sv.viewer_id = ?", userID).
		Order("u.id ASC, s.created_at ASC").
		Scan(&rows).Error
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to build feed"})
	}

	// Response structs
	type storyItem struct {
		ID        uint      `json:"id"`
		MediaURL  string    `json:"media_url"`
		MediaType string    `json:"media_type"`
		CreatedAt time.Time `json:"created_at"`
		Seen      bool      `json:"seen"`
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
			CreatedAt: r.CreatedAt,
			Seen:      seen,
		})
	}

	// Preserve ordering by iterating through rows
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

// ---------- Get user stories: GET /stories/user/:id ----------
func GetUserStories(c echo.Context) error {
	uidStr := c.Param("id")
	targetID, err := strconv.Atoi(uidStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid user id"})
	}

	// Check user exists
	var count int64
	if err := config.DB.Model(&models.User{}).
		Where("id = ?", targetID).
		Count(&count).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "database error"})
	}
	if count == 0 {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
	}

	var stories []models.Story
	if err := config.DB.Where("user_id = ? AND expires_at > ?", targetID, time.Now()).
		Order("created_at desc").
		Find(&stories).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "database error"})
	}

	return c.JSON(http.StatusOK, stories)
}

// ---------- Add story: POST /stories ----------
type addStoryReq struct {
	MediaURL   string `json:"media_url"`
	MediaType  string `json:"media_type"`
	TTLMinutes int    `json:"ttl_minutes"`
}

func AddStory(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	var req addStoryReq
	if err := c.Bind(&req); err != nil || req.MediaURL == "" || req.MediaType == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}

	// TTL (default 24h, max 24h)
	if req.TTLMinutes <= 0 || req.TTLMinutes > 1440 {
		req.TTLMinutes = 1440
	}
	ttl := time.Duration(req.TTLMinutes) * time.Minute

	story := models.Story{
		UserID:    userID,
		MediaURL:  req.MediaURL,
		MediaType: req.MediaType,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(ttl),
	}

	if err := config.DB.Create(&story).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "database error"})
	}
	return c.JSON(http.StatusCreated, story)
}

// ---------- Delete story: DELETE /stories/:id ----------
func DeleteStory(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid story id"})
	}

	var story models.Story
	if err := config.DB.First(&story, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "story not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	if err := config.DB.Delete(&story).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "delete failed"})
	}

	return c.NoContent(http.StatusNoContent)
}

// ---------- Mark story as viewed: POST /stories/:id/view ----------
func ViewStory(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	idStr := c.Param("id")
	storyID, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid story id"})
	}

	// Ensure story exists and is active
	var story models.Story
	if err := config.DB.First(&story, "id = ? AND expires_at > ?", storyID, time.Now()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "story not found or expired"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	// Insert into story_views (ignore duplicate views)
	view := models.StoryView{
		StoryID:  uint(storyID),
		ViewerID: userID,
		ViewedAt: time.Now(),
	}
	if err := config.DB.Create(&view).Error; err != nil {
		return c.JSON(http.StatusOK, echo.Map{"message": "already viewed"})
	}

	return c.JSON(http.StatusCreated, echo.Map{"message": "view recorded"})
}

// ---------- Get views of a story: GET /stories/:id/views ----------
func GetStoryViews(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	idStr := c.Param("id")
	storyID, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid story id"})
	}

	// Ensure story exists and belongs to logged-in user (privacy check)
	var story models.Story
	if err := config.DB.First(&story, "id = ?", storyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"error": "story not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}
	if story.UserID != userID {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "not your story"})
	}

	// Fetch views with viewer info
	var views []struct {
		ID         uint      `json:"id"`
		ViewerID   uint      `json:"viewer_id"`
		Username   string    `json:"username"`
		ProfilePic *string   `json:"profile_pic"`
		ViewedAt   time.Time `json:"viewed_at"`
	}
	if err := config.DB.Table("story_views").
		Select("story_views.id, story_views.viewer_id, users.username, users.profile_pic, story_views.viewed_at").
		Joins("JOIN users ON users.id = story_views.viewer_id").
		Where("story_views.story_id = ?", storyID).
		Order("story_views.viewed_at desc").
		Find(&views).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	// Count total views
	totalViews := len(views)

	return c.JSON(http.StatusOK, echo.Map{
		"story_id":    storyID,
		"total_views": totalViews,
		"views":       views,
	})
}
