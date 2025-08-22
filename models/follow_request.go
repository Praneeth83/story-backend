package models

type FollowRequest struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	FollowerID uint `gorm:"not null;uniqueIndex:idx_follow_requests_unique" json:"follower_id"` // who follows
	FolloweeID uint `gorm:"not null;uniqueIndex:idx_follow_requests_unique" json:"followee_id"` // who is being followed

	// -------- Relations --------
	Follower User `gorm:"foreignKey:FollowerID" json:"follower"` // the follower user
	Followee User `gorm:"foreignKey:FolloweeID" json:"followee"` // the followed user
}
