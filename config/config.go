package config

import (
	"log"
	"os"
)

var (
	DSN       string
	JWTSecret string
)

func LoadConfig() {
	// Load DB DSN
	DSN = os.Getenv("DB_DSN")
	if DSN == "" {
		// Local default DB connection
		DSN = "host=localhost user=postgres password=praneeth dbname=lol port=5432 sslmode=disable"
	}

	// Load JWT secret
	JWTSecret = os.Getenv("JWT_SECRET")
	if JWTSecret == "" {
		// Default for local dev only — never use in production
		JWTSecret = "supersecretkey"
	}

	log.Println("✅ Config loaded")
}
