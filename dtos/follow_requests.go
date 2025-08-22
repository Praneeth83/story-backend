package dtos

import "story-backend/models"

func NewFollowFromRequest(fr models.FollowRequest) models.Follow {
	return models.Follow{
		FollowerID: fr.FollowerID,
		FolloweeID: fr.FolloweeID,
	}
}
