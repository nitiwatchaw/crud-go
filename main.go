package main

import (
	routes "nitiwat/routes"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use()

	routes.AuthRouter(router)
	routes.UserRouter(router)
	routes.TodoRouter(router)

	router.Run(":" + port)
}
