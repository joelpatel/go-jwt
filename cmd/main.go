package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joelpatel/go-jwt/routes"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Access granted for api-1."})
	})

	router.GET("/api-2", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"message": "Access granted for api-2."})
	})

	router.Run(":" + port)
}
