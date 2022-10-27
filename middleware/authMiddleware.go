package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/joelpatel/go-jwt/helpers"
)

func Authenticate() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientToken := ctx.Request.Header.Get("token")
		if clientToken == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "no authorization header provided"})
			ctx.Abort() // pending handlers won't be called
			return
		}

		claims, errMsg := helpers.ValidateToken(clientToken)
		if errMsg != "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": errMsg})
			ctx.Abort()
			return
		}

		ctx.Set("email", claims.Email)
		ctx.Set("firt_name", claims.FirstName)
		ctx.Set("user_type", claims.UserType)

	}
}
