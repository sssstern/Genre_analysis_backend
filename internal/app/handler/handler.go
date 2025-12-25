package handler

import (
	"lab4/internal/app/middleware"
	"lab4/internal/app/repository"
	"net/http" // <-- Обязательный импорт
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository  *repository.Repository
	MinioClient *minio.Client
	BucketName  string
	RedisClient *redis.Client
	SecretKey   string
	HostName    string
	JWTDur      time.Duration

	// 🔥 КРИТИЧЕСКОЕ ИЗМЕНЕНИЕ 1: Добавление полей
	DjangoServiceURL  string       // URL для вызова Django-сервиса
	HTTPClient        *http.Client // HTTP-клиент для запросов (ИСТОЧНИК ПАНИКИ, если nil)
	CallbackSecretKey string
}

// 🔥 КРИТИЧЕСКОЕ ИЗМЕНЕНИЕ 2: Обновление сигнатуры и инициализация
func NewHandler(r *repository.Repository, minioClient *minio.Client, bucketName string, rdb *redis.Client, secretKey, hostName, djangoServiceURL string, jwtDur time.Duration, CallbackSecretKey string) *Handler {

	// ИНИЦИАЛИЗАЦИЯ HTTP-КЛИЕНТА
	httpClient := &http.Client{Timeout: 15 * time.Second}

	return &Handler{
		Repository:  r,
		MinioClient: minioClient,
		BucketName:  bucketName,
		RedisClient: rdb,
		SecretKey:   secretKey,
		HostName:    hostName,
		JWTDur:      jwtDur,

		// ИНИЦИАЛИЗАЦИЯ НОВЫХ ПОЛЕЙ
		DjangoServiceURL:  djangoServiceURL,
		HTTPClient:        httpClient, // <-- Теперь не nil!
		CallbackSecretKey: CallbackSecretKey,
	}
}

func (h *Handler) RegisterStatic(router *gin.Engine) {
	router.Static("/img", "./resources/img")
}

func (h *Handler) errorHandler(ctx *gin.Context, errorStatusCode int, err error) {
	logrus.Error(err.Error())
	ctx.JSON(errorStatusCode, gin.H{
		"message": err.Error(),
	})
}

func (h *Handler) successResponse(ctx *gin.Context, data interface{}) {
	ctx.JSON(200, gin.H{
		"data": data,
	})
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func (h *Handler) Ping(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func (h *Handler) RegisterHandler(router *gin.Engine) {
	api := router.Group("/api/v1")
	{
		api.POST("/user/register", h.RegisterUser)
		api.POST("/user/login", h.LoginUser)

		api.GET("/genres", h.GetGenres)
		api.GET("/genres/:id", h.GetGenre)

		api.GET("/text-analysis-request/icon", h.GetCurrentAnalysis)

		//api.GET("/ping", h.Ping)
		api.PUT("/text-analysis-request/update-analysis", h.UpdateAnalysisResult)
	}

	auth := router.Group("/api/v1")
	auth.Use(middleware.AuthMiddleware(h.SecretKey, h.RedisClient))
	{
		auth.POST("/user/logout", h.LogoutUser)
		auth.GET("/user/profile", h.GetProfile)
		auth.PUT("/user/profile", h.UpdateProfile)

		auth.GET("/text-analysis-request", h.GetAnalysisRequests)
		auth.GET("/text-analysis-requests/:id", h.GetAnalysisRequestByID)
		auth.PUT("/text-analysis-requests/:id", h.UpdateAnalysisRequest)
		auth.PUT("/text-analysis-requests/:id/form", h.FormAnalysisRequest)
		auth.DELETE("/text-analysis-requests/:id", h.DeleteAnalysisRequest)

		auth.POST("/genres/add-to-analysis/:id", h.AddGenreToAnalysis)
		auth.PUT("/analysis-genres/:id", h.UpdateAnalysisGenre)
		auth.DELETE("/analysis-genres/:id", h.RemoveGenreFromAnalysis)
	}

	moderator := auth.Group("")
	moderator.Use(middleware.RequireModerator())
	{
		moderator.POST("/genres", h.CreateGenre)
		moderator.PUT("/genres/:id", h.UpdateGenre)
		moderator.DELETE("/genres/:id", h.DeleteGenre)
		moderator.POST("/genres/:id/image", h.UploadGenreImage)

		moderator.PUT("/text-analysis-requests/:id/process", h.ProcessAnalysisRequest)
	}
}
