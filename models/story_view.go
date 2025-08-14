package models

import "time"

type StoryView struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	StoryID  uint      `gorm:"index:idx_story_viewer,unique;not null" json:"story_id"`
	ViewerID uint      `gorm:"index:idx_story_viewer,unique;not null" json:"viewer_id"`
	ViewedAt time.Time `gorm:"autoCreateTime" json:"viewed_at"`
}
