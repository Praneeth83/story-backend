package dtos

import "story-backend/models"

// -------------------- Error --------------------
type Error struct {
	Error string `json:"error"`
}

func ErrorResponse(msg string) Error {
	return Error{Error: msg}
}

// -------------------- User Response --------------------
type UserResponse struct {
	ID         uint    `json:"id"`
	Username   string  `json:"username"`
	Email      string  `json:"email"`
	ProfilePic *string `json:"profile_pic"`
	Type       string  `json:"type"`
}

func ToUserResponse(u models.User) UserResponse {
	return UserResponse{
		ID:         u.ID,
		Username:   u.Username,
		Email:      u.Email,
		ProfilePic: u.ProfilePic,
		Type:       u.Type,
	}
}

// -------------------- Auth --------------------
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"`
}

func NewAuthResponse(user models.User, token string) AuthResponse {
	return AuthResponse{
		User:  ToUserResponse(user),
		Token: token,
	}
}

func NewAuthAutoResponse(user models.User, token string) AuthResponse {
	// same as normal auth, just naming distinction
	return NewAuthResponse(user, token)
}

// -------------------- Me --------------------
type MeResponse struct {
	User UserResponse `json:"user"`
}

func NewMeResponse(user models.User) MeResponse {
	return MeResponse{User: ToUserResponse(user)}
}

// -------------------- Toggle Type --------------------
type ToggleResponse struct {
	Type string `json:"type"`
}

func NewToggleResponse(t string) ToggleResponse {
	return ToggleResponse{Type: t}
}
