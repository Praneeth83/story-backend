package models

import "time"

type User struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username   string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email      string    `gorm:"uniqueIndex;size:120;not null" json:"email"`
	Password   string    `gorm:"not null" json:"-"`
	ProfilePic *string   `gorm:"type:text" json:"profile_pic,omitempty"` // Nullable
	Type       string    `gorm:"type:text;default:'public'" json:"type"` // "public" or "private"
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}
