package internal

import (
	"fmt"
	"time"

	"story-backend/config"
	"story-backend/models"
)

func DeleteExpiredStories() {
	result := config.DB.Where("expires_at <= ?", time.Now()).Delete(&models.Story{})
	if result.Error == nil && result.RowsAffected > 0 {
		fmt.Printf("🗑 Deleted %d expired stories\n", result.RowsAffected)
	}
}
