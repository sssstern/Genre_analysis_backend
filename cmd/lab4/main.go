package main

import (
	"fmt"
	"time"

	"context"
	"lab4/internal/app/config"
	"lab4/internal/app/dsn"
	"lab4/internal/app/handler"
	"lab4/internal/app/middleware"
	"lab4/internal/app/repository"
	"lab4/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "lab4/docs"

	//"github.com/gin-contrib/cors"
	"github.com/redis/go-redis/v9"
)

// @title Анализ Принадлежности Текста к Жанру
// @version 1.0
// @description Бэкенд сервис для анализа текстовых заявок и определения жанра.
// @host localhost:8082
// @BasePath /api/v1
// @schemes http
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Используется для запросов через Insomnia/Postman: "Bearer <JWT>"
// @securityDefinitions.cookie SessionCookie
// @name session_token
// @description Используется для запросов из браузера.
// @in cookie
func main() {
	router := gin.Default()
	router.Use(middleware.CORS())
	//router.Use(cors.Default())
	/*router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://sssstern.github.io", // ← ЗАМЕНИ на свой настоящий GitHub Pages URL
			"http://localhost:5173",      // ← для локальной разработки (Vite)
			"http://127.0.0.1:5173",
			"https://192.168.1.48:8443",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true, // ← КРИТИЧНО! Без этого куки и сессия не будут передаваться
		MaxAge:           12 * time.Hour,
	}))*/
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Addr,
		Password: conf.Redis.Password,
		DB:       conf.Redis.DB,
	})

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		logrus.Fatalf("Ошибка подключения к Redis: %v", err)
	}
	logrus.Info("Успешное подключение к Redis.")

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString, rdb)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		logrus.Fatal("Failed to load config: ", err)
	}

	minioClient, err := config.NewMinioClient(cfg.Minio)
	if err != nil {
		logrus.Fatal("Failed to initialize Minio client: ", err)
	}

	jwtDuration, err := time.ParseDuration(conf.JWT.ExpiresIn)
	if err != nil {
		logrus.Fatalf("Invalid JWT expiration duration in config: %v", err)
	}

	djangoServiceURL := "http://127.0.0.1:8000/asyncapi/v1/calculate-text-genre-probability"
	CallbackSecretKey := "GenreKey"

	hand := handler.NewHandler(
		rep,
		minioClient,
		conf.Minio.BucketName,
		rdb,
		conf.JWT.SecretKey,
		conf.ServiceHost,
		djangoServiceURL, // <-- Добавлен восьмой аргумент
		jwtDuration,
		CallbackSecretKey,
	)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()
}
