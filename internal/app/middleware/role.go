package middleware

import (
	"net/http"

	"lab4/internal/app/ds"

	"github.com/gin-gonic/gin"
)

func RequireModerator() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if GetRole(ctx) != ds.RoleModerator {
			ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "Доступ запрещен. Требуются права модератора."})
			return
		}
		ctx.Next()
	}
}

func RequireAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := GetRole(ctx)
		if role != ds.RoleCreator && role != ds.RoleModerator {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Требуется авторизация."})
			return
		}
		ctx.Next()
	}
}

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "*")
		c.Header("Access-Control-Expose-Headers", "Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
