package main

import (
	"fmt"

	"lab3/internal/app/config"
	"lab3/internal/app/dsn"
	"lab3/internal/app/handler"
	"lab3/internal/app/repository"
	"lab3/internal/pkg"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	router := gin.Default()
	conf, err := config.NewConfig()
	if err != nil {
		logrus.Fatalf("error loading config: %v", err)
	}

	postgresString := dsn.FromEnv()
	fmt.Println(postgresString)

	rep, errRep := repository.New(postgresString)
	if errRep != nil {
		logrus.Fatalf("error initializing repository: %v", errRep)
	}

	cfg, err := config.NewConfig()
	if err != nil {
		logrus.Fatal("Failed to load config: ", err)
	}

	// Инициализация Minio клиента
	minioClient, err := config.NewMinioClient(cfg.Minio)
	if err != nil {
		logrus.Fatal("Failed to initialize Minio client: ", err)
	}

	hand := handler.NewHandler(rep, minioClient, cfg.Minio.BucketName)

	// Инициализация обработчика
	application := pkg.NewApp(conf, router, hand)
	application.RunApp()

}
