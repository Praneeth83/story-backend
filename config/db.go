package config

import (
	"fmt"
	"log"
	"story-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	DB, err = gorm.Open(postgres.Open(DSN), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	if err := DB.AutoMigrate(&models.User{}, &models.Story{}, &models.StoryView{}, &models.Follow{}, &models.FollowRequest{}); err != nil {
		log.Fatal("AutoMigration failed:", err)
	}

	fmt.Println("✅ Database connection successful & migrated")
}
