package config

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func NewMinioClient(cfg MinioConfig) (*minio.Client, error) {
	log.Printf("Initializing Minio client with endpoint: %s", cfg.Endpoint)

	// Создаем клиент Minio
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Minio client: %v", err)
	}

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Проверяем подключение списком бакетов
	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Minio: %v", err)
	}

	// Проверяем существует ли бакет
	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		return nil, fmt.Errorf("bucket '%s' does not exist. Please create it first in Minio console", cfg.BucketName)
	}

	log.Printf(" Minio client initialized successfully with bucket '%s'", cfg.BucketName)
	return minioClient, nil
}
