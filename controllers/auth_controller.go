package controllers

import (
	"net/http"
	"strings"

	"story-backend/config"
	"story-backend/models"
	"story-backend/utils"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type signupReq struct {
	Username   string  `json:"username"`
	Email      string  `json:"email"`
	Password   string  `json:"password"`
	ProfilePic *string `json:"profile_pic"` // Optional
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Signup creates a new user account
func Signup(c echo.Context) error {
	var req signupReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	if req.Username == "" || req.Email == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid fields"})
	}

	hashed, err := utils.HashPassword(req.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "hash error"})
	}

	user := models.User{
		Username:   req.Username,
		Email:      req.Email,
		Password:   hashed,
		ProfilePic: req.ProfilePic,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "user exists or bad data"})
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "token error"})
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"user":  userResponse(user),
		"token": token,
	})
}

// Login authenticates user with token-first logic
func Login(c echo.Context) error {
	// 1. Try auto-login with token
	authHeader := c.Request().Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := utils.ParseToken(tokenStr)
		if err == nil {
			if userID, err := utils.ExtractUserID(claims); err == nil {
				var user models.User
				if err := config.DB.First(&user, userID).Error; err == nil {
					return c.JSON(http.StatusOK, echo.Map{
						"user":  userResponse(user),
						"token": tokenStr,
						"auto":  true,
					})
				}
			}
		}
	}

	// 2. Fallback to email/password login
	var req loginReq
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid input"})
	}

	var user models.User
	if err := config.DB.Where("email = ?", strings.ToLower(req.Email)).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid credentials"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	if err := utils.CheckPassword(user.Password, req.Password); err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid credentials"})
	}

	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "token error"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"user":  userResponse(user),
		"token": token,
		"auto":  false,
	})
}

// Me returns the logged-in user's info (middleware provides user_id)
func Me(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
	}

	return c.JSON(http.StatusOK, echo.Map{"user": userResponse(user)})
}

// ToggleAccountType flips between public and private

func ToggleAccountType(c echo.Context) error {
	userID, err := utils.GetUserID(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"error": "user not found"})
	}

	if user.Type == "private" {
		// Switch to public
		user.Type = "public"

		// Auto-accept all follow requests
		var requests []models.FollowRequest
		if err := config.DB.Where("followee_id = ?", user.ID).Find(&requests).Error; err == nil {
			for _, req := range requests {
				follow := models.Follow{
					FollowerID: req.FollowerID,
					FolloweeID: req.FolloweeID,
				}
				// ignore if already exists
				config.DB.FirstOrCreate(&follow, follow)
				config.DB.Delete(&req) // clean up request
			}
		}
	} else {
		// Switch to private
		user.Type = "private"
	}

	if err := config.DB.Save(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "db error"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"message": "account type toggled successfully",
		"type":    user.Type,
	})
}

// -------------------- Helpers --------------------
func userResponse(user models.User) echo.Map {
	return echo.Map{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"profile_pic": user.ProfilePic,
	}
}
