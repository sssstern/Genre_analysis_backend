package handler

import (
	"lab3/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository  *repository.Repository
	MinioClient *minio.Client
	BucketName  string
}

func NewHandler(r *repository.Repository, minioClient *minio.Client, bucketName string) *Handler {
	return &Handler{
		Repository:  r,
		MinioClient: minioClient,
		BucketName:  bucketName,
	}
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.Static("/img", "./resources/img")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"status":  "error",
		"message": err.Error(),
	})
}

func (h *Handler) successResponse(ctx *gin.Context, data interface{}) {
	ctx.JSON(200, gin.H{
		"status": "success",
		"data":   data,
	})
}

// Временная функция для имитации авторизации
func (h *Handler) getCurrentUserID() int {
	return 1
}

func (h *Handler) getCurrentModeratorID() int {
	return 2
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		// Жанры
		api.GET("/genres", h.GetGenres)
		api.GET("/genres/:id", h.GetGenre)
		api.POST("/genres", h.CreateGenre)
		api.PUT("/genres/:id", h.UpdateGenre)
		api.DELETE("/genres/:id", h.DeleteGenre)
		api.POST("/genres/:id/image", h.UploadGenreImage)
		api.POST("/genres/add-to-analysis/:id", h.AddGenreToAnalysis)

		// Заявки
		api.GET("/analysis-request/icon", h.GetCurrentAnalysis)
		api.GET("/analysis-request", h.GetAnalysisRequests)
		api.GET("/analysis-requests/:id", h.GetAnalysisRequestByID)
		api.PUT("/analysis-requests/:id", h.UpdateAnalysisRequest)
		api.PUT("/analysis-requests/:id/form", h.FormAnalysisRequest)
		api.DELETE("/analysis-requests/:id", h.DeleteAnalysisRequest)
		api.PUT("/analysis-requests/:id/process", h.ProcessAnalysisRequest)

		// m-m
		api.DELETE("/analysis-genre/:id", h.RemoveGenreFromAnalysis)
		api.PUT("/analysis-genre/:id", h.UpdateAnalysisGenre)

		// Пользователи
		api.POST("/user/register", h.RegisterUser)
		api.GET("/user/profile", h.GetProfile)
		api.PUT("/user/profile", h.UpdateProfile)
		api.POST("/user/login", h.LoginUser)
		api.POST("/user/logout", h.LogoutUser)
	}
}
