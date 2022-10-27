package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/joelpatel/go-jwt/controllers"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("/users/signup", controllers.Signup())
	router.POST("/users/login", controllers.Login())
}
