package models

import "time"

type Story struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	MediaURL  string    `gorm:"type:text;not null" json:"media_url"`
	MediaType string    `gorm:"size:20;not null" json:"media_type"` // "image" | "video"
	CreatedAt time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	ExpiresAt time.Time `gorm:"index;not null" json:"expires_at"`

	// -------- Relations --------
	User  User        `gorm:"foreignKey:UserID" json:"user"`
	Views []StoryView `gorm:"foreignKey:StoryID" json:"views,omitempty"`
}
