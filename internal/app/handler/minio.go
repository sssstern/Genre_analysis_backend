package handler

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
)

func (h *Handler) generateImageName(originalName string, genreID int) string {
	ext := filepath.Ext(originalName)

	baseName := strings.TrimSuffix(strings.ToLower(originalName), ext)

	cleanBaseName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, baseName)

	cleanBaseName = strings.Join(strings.FieldsFunc(cleanBaseName, func(r rune) bool {
		return r == '_'
	}), "_")

	if len(cleanBaseName) > 20 {
		cleanBaseName = cleanBaseName[:20]
	}
	cleanBaseName = strings.Trim(cleanBaseName, "_")

	if cleanBaseName == "" {
		cleanBaseName = "file"
	}

	return fmt.Sprintf("%s_genre_%d_pic%s", cleanBaseName, genreID, ext)
}

func (h *Handler) deleteImageFromMinio(imageURL string) error {
	if imageURL == "" || h.MinioClient == nil {
		return nil
	}

	objectName := filepath.Base(imageURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := h.MinioClient.RemoveObject(ctx, h.BucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete image from Minio: %v", err)
	}

	return nil
}

func (h *Handler) uploadImageToMinio(file *multipart.FileHeader, objectName string) (string, error) {
	if h.MinioClient == nil {
		return "", fmt.Errorf("minio client not available")
	}

	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %v", err)
	}
	defer src.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = h.MinioClient.PutObject(ctx, h.BucketName, objectName, src, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image to Minio: %v", err)
	}

	imageURL := fmt.Sprintf("/%s/%s", h.BucketName, objectName)
	return imageURL, nil
}

func (h *Handler) isValidImage(file *multipart.FileHeader) bool {
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	contentType := file.Header.Get("Content-Type")
	return allowedTypes[contentType]
}

func (h *Handler) checkMinioConnection() bool {
	if h.MinioClient == nil {
		logrus.Warn("Minio client is nil")
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := h.MinioClient.BucketExists(ctx, h.BucketName)
	if err != nil {
		logrus.Errorf("Minio connection check failed: %v", err)
		return false
	}

	logrus.Info("Minio connection check: OK")
	return true
}
