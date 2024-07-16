package routes

import (
	"nitiwat/controllers"
	"nitiwat/middleware"

	"github.com/gin-gonic/gin"
)

func DeletedRouter(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/deleted", controllers.GetAllDeleted())
	incomingRoutes.GET("/deleted/:del_id", controllers.GetDeletedById())

}
