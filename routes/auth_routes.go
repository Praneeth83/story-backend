package routes

import (
	"story-backend/controllers"
	"story-backend/middleware"

	"github.com/labstack/echo/v4"
)

func AuthRoutes(e *echo.Echo) {
	auth := e.Group("/auth")

	// Public routes
	auth.POST("/signup", controllers.Signup)
	auth.POST("/login", controllers.Login)

	// Protected route to get current logged-in user
	auth.GET("/me", controllers.Me, middleware.JWTAuth())
	auth.PATCH("/toggle", controllers.ToggleAccountType, middleware.JWTAuth())

}
