package models

import "time"

type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	Caption   string    `gorm:"type:text" json:"caption"`
	MediaURL  string    `gorm:"type:text;not null" json:"media_url"`
	MediaType string    `gorm:"size:20;not null" json:"media_type"` // "image" | "video"
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// -------- Relations --------
	User User `gorm:"foreignKey:UserID" json:"user"`
}
