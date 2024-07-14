package routes

import (
	"nitiwat/controllers"
	"nitiwat/middleware"

	"github.com/gin-gonic/gin"
)

func TodoRouter(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/todos", controllers.GetTodo())
	incomingRoutes.GET("/todos/:todo_id", controllers.GetTodoById())
	incomingRoutes.GET("/todos-user/:user_id", controllers.GetTodoByUser())
	incomingRoutes.POST("/todos", controllers.AddTodo())
	incomingRoutes.PUT("/todos/:todo_id", controllers.UpdateCheck())
	incomingRoutes.PUT("/todos-update/:todo_id", controllers.UpdateEditTodo())
	incomingRoutes.GET("todos-active", controllers.CheckALlTodoActive())
	incomingRoutes.GET("todos-find", controllers.FindQuery())
	// incomingRoutes.GET("/todos/:todo_id", controller.GetTodo())
	// incomingRoutes.PUT("/todos/:todo_id", controller.UpdateTodo())
	incomingRoutes.DELETE("/todos/:todo_id", controllers.DeleteTodo())
}
