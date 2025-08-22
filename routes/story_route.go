package routes

import (
	"story-backend/controllers"
	"story-backend/middleware"

	"github.com/labstack/echo/v4"
)

func StoryRoutes(e *echo.Echo) {
	stories := e.Group("/stories", middleware.JWTAuth())
	stories.POST("/add", controllers.AddStory)
	stories.GET("/feed", controllers.GetStoriesFeed)
	stories.GET("/user/:id", controllers.GetUserStories)
	stories.DELETE("/:id", controllers.DeleteStory)
	stories.POST("/:id/view", controllers.ViewStory)
	stories.GET("/:id/views", controllers.GetStoryViews)

}
