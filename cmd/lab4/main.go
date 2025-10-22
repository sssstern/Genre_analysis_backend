package main

import (
	"fmt"
	"time"

	"context"
	"lab4/internal/app/config"
	"lab4/internal/app/dsn"
	"lab4/internal/app/handler"
	"lab4/internal/app/repository"
	"lab4/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	_ "lab4/docs"

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

	hand := handler.NewHandler(
		rep,
		minioClient,
		conf.Minio.BucketName,
		rdb,
		conf.JWT.SecretKey,
		conf.ServiceHost,
		jwtDuration,
	)

	application := pkg.NewApp(conf, router, hand)
	application.RunApp()

}
