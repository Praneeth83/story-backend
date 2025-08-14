package routes

import (
	"story-backend/controllers"
	"story-backend/middleware"

	"github.com/labstack/echo/v4"
)

func FollowRoutes(e *echo.Echo) {
	follow := e.Group("/follow", middleware.JWTAuth())

	follow.POST("/:id", controllers.FollowUser)            // Follow a user
	follow.DELETE("/:id", controllers.UnfollowUser)        // Unfollow a user
	follow.GET("/following/:id", controllers.GetFollowing) // Who this user follows
	follow.GET("/followers/:id", controllers.GetFollowers) // Who follows this user
}
