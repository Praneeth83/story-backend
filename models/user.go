package models

import "time"

type User struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username   string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email      string    `gorm:"uniqueIndex;size:120;not null" json:"email"`
	Password   string    `gorm:"not null" json:"-"`
	ProfilePic *string   `gorm:"type:text" json:"profile_pic,omitempty"`
	Type       string    `gorm:"type:text;default:'public'" json:"type"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`

	// -------- Relations --------
	Posts      []Post      `gorm:"foreignKey:UserID" json:"posts,omitempty"`
	Stories    []Story     `gorm:"foreignKey:UserID" json:"stories,omitempty"`
	Followers  []Follow    `gorm:"foreignKey:FolloweeID" json:"followers,omitempty"`
	Following  []Follow    `gorm:"foreignKey:FollowerID" json:"following,omitempty"`
	StoryViews []StoryView `gorm:"foreignKey:ViewerID" json:"story_views,omitempty"`

	// Only received requests
	FollowRequestsReceived []FollowRequest `gorm:"foreignKey:FolloweeID" json:"follow_requests_received,omitempty"`
}
