package routes

import (
	"nitiwat/controllers"
	"nitiwat/middleware"

	"github.com/gin-gonic/gin"
)

func UserRouter(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controllers.GetUsers())
	incomingRoutes.GET("/users/:user_id", controllers.GetUser())
	incomingRoutes.DELETE("/users/:user_id", controllers.DeleteUser())
}
