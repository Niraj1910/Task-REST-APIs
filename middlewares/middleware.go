package middlewares

import (
	"net/http"
	"strings"

	"github.com/Niraj1910/Task-REST-APIs/auth"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(ctx *gin.Context) {

	// check the token in the cookie first
	tokenString, err := ctx.Cookie("token")
	if tokenString == "" || err != nil {
		// fallback if token not found in the cookie
		authHeader := ctx.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.Split(authHeader, " ")
			if len(parts) == 2 || parts[0] == "Bearer" {
				tokenString = parts[1]
			}
		}
	}

	if tokenString == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"error": "Authentication required (no token in cookie or header)",
		})
		ctx.Abort()
		return
	}

	claims, err := auth.VerifyToken(tokenString)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired  token", "details": err.Error()})
		ctx.Abort()
		return
	}

	ctx.Set("user_id", claims.ID)
	ctx.Next()

}
