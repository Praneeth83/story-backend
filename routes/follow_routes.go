package routes

import (
	"story-backend/controllers"
	"story-backend/middleware"

	"github.com/labstack/echo/v4"
)

func FollowRoutes(e *echo.Echo) {
	follow := e.Group("/follow", middleware.JWTAuth())

	follow.POST("/:identifier", controllers.FollowUser)    // Follow a user (by id or username)
	follow.DELETE("/:id", controllers.UnfollowUser)        // Unfollow a user
	follow.GET("/following/:id", controllers.GetFollowing) // Who this user follows
	follow.GET("/followers/:id", controllers.GetFollowers) // Who follows this user

	// NEW for private accounts:
	follow.GET("/requests", controllers.GetFollowRequests)          // list requests
	follow.POST("/requests/:id", controllers.AcceptFollowRequest)   // accept request
	follow.DELETE("/requests/:id", controllers.RejectFollowRequest) // reject request

}
