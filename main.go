package main

import (
	"log"
	"story-backend/config"
	"story-backend/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables & configuration
	config.LoadConfig()

	// Connect to DB
	config.ConnectDB()

	// Init Echo
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	routes.AuthRoutes(e)
	routes.StoryRoutes(e)
	routes.FollowRoutes(e)
	// Start server
	log.Println("🚀 Server started at :8080")
	if err := e.Start("192.168.0.111:8080"); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
