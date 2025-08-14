package models

type Follow struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	FollowerID uint `json:"follower_id"`
	FolloweeID uint `json:"followee_id"`
}
