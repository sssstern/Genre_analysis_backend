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
