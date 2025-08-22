package dtos

import "story-backend/models"

// -------------------- Signup --------------------
type SignupRequest struct {
	Username   string  `json:"username"`
	Email      string  `json:"email"`
	Password   string  `json:"password"`
	ProfilePic *string `json:"profile_pic"`
}

func (r SignupRequest) ToUser(hashed string) models.User {
	return models.User{
		Username:   r.Username,
		Email:      r.Email,
		Password:   hashed,
		ProfilePic: r.ProfilePic,
		Type:       "public", // default
	}
}

// -------------------- Login --------------------
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
