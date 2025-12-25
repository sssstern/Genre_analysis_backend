package middleware

import (
	"context"
	"errors"
	"net/http"

	"lab4/internal/app/ds" // Импорт ds для доступа к UserRole
	"lab4/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func AuthMiddleware(secretKey string, rdb *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := service.ExtractToken(ctx)
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Неавторизован. Отсутствует токен."})
			return
		}

		val, err := rdb.Get(context.Background(), tokenString).Result()
		if err == nil && val == "blacklist" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Токен недействителен (выход из системы)."})
			return
		}
		if err != nil && !errors.Is(err, redis.Nil) {
			logrus.Error("Redis Error in AuthMiddleware:", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		claims, err := service.ParseJWT(tokenString, secretKey)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Недействительный токен."})
			return
		}

		ctx.Set("userID", claims.UserID)
		ctx.Set("userRole", claims.Role)
		ctx.Set("tokenString", tokenString)

		ctx.Next()
	}
}

func GetUserID(ctx *gin.Context) int {
	userID, ok := ctx.Get("userID")
	if !ok {
		return 0
	}
	return userID.(int)
}

func IsModerator(ctx *gin.Context) bool {
	isModerator, ok := ctx.Get("isModerator")
	if !ok {
		return false
	}
	return isModerator.(bool)
}

func GetTokenString(ctx *gin.Context) string {
	token, ok := ctx.Get("tokenString")
	if !ok {
		return ""
	}
	return token.(string)
}

func GetRole(ctx *gin.Context) ds.UserRole {
	role, ok := ctx.Get("userRole")
	if !ok {
		return ds.RoleGuest
	}
	return role.(ds.UserRole)
}
