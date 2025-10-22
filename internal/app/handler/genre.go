// handler/genre.go
package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"lab3/internal/app/ds"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handler) GetGenres(ctx *gin.Context) {
	searchQuery := ctx.Query("searchbygenrename")

	var genres []ds.GenreDTO
	var err error

	if searchQuery == "" {
		genres, err = h.Repository.GetGenres()
	} else {
		genres, err = h.Repository.GetGenresByTitle(searchQuery)
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successResponse(ctx, genres)
}

func (h *Handler) GetGenre(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	genreDTO, err := h.Repository.GetGenre(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	h.successResponse(ctx, genreDTO)
}

func (h *Handler) CreateGenre(ctx *gin.Context) {
	var genreDTO ds.GenreDTO
	if err := ctx.BindJSON(&genreDTO); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	createdGenreDTO, err := h.Repository.CreateGenre(genreDTO)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"data": createdGenreDTO,
	})
}

func (h *Handler) UpdateGenre(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var genreUpdates ds.UpdateGenreDTO
	if err := ctx.BindJSON(&genreUpdates); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	updatedGenreDTO, err := h.Repository.UpdateGenre(id, genreUpdates)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successResponse(ctx, updatedGenreDTO)
}

func (h *Handler) DeleteGenre(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	genreDTO, err := h.Repository.GetGenre(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if genreDTO.GenreImageURL != "" && h.MinioClient != nil {
		err = h.deleteImageFromMinio(genreDTO.GenreImageURL)
		if err != nil {
			logrus.Errorf("Failed to delete image from Minio for genre %d: %v", id, err)
		}
	}

	err = h.Repository.DeleteGenre(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (h *Handler) UploadGenreImage(ctx *gin.Context) {
	if h.MinioClient == nil {
		h.errorHandler(ctx, http.StatusServiceUnavailable,
			fmt.Errorf("image storage service not configured"))
		return
	}

	if !h.checkMinioConnection() {
		h.errorHandler(ctx, http.StatusServiceUnavailable,
			fmt.Errorf("image storage service temporarily unavailable. Please try again later"))
		return
	}

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("invalid genre ID: %v", err))
		return
	}

	genreDTO, err := h.Repository.GetGenre(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("genre not found: %v", err))
		return
	}

	file, err := ctx.FormFile("image")
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("image file is required: %v", err))
		return
	}

	if !h.isValidImage(file) {
		h.errorHandler(ctx, http.StatusBadRequest,
			fmt.Errorf("invalid image format. Allowed: JPEG, PNG, GIF, WebP"))
		return
	}

	if file.Size > 5*1024*1024 {
		h.errorHandler(ctx, http.StatusBadRequest,
			fmt.Errorf("image size too large. Maximum 5MB allowed"))
		return
	}

	if genreDTO.GenreImageURL != "" {
		err = h.deleteImageFromMinio(genreDTO.GenreImageURL)
		if err != nil {
			logrus.Warnf("Failed to delete old image from Minio: %v", err)
		}
	}
	objectName := h.generateImageName(file.Filename, id)

	imageURL, err := h.uploadImageToMinio(file, objectName)
	if err != nil {
		logrus.Errorf("Failed to upload image to Minio: %v", err)
		h.errorHandler(ctx, http.StatusInternalServerError,
			fmt.Errorf("failed to upload image: %v", err))
		return
	}

	updatedGenreDTO, err := h.Repository.UpdateGenreImage(id, imageURL)
	if err != nil {
		// Если не удалось обновить БД, удаляем загруженное изображение
		logrus.Errorf("Failed to update genre in database, rolling back image upload")
		h.deleteImageFromMinio(imageURL)
		h.errorHandler(ctx, http.StatusInternalServerError,
			fmt.Errorf("failed to update genre: %v", err))
		return
	}

	logrus.Infof("Successfully updated genre %d with new image: %s", id, imageURL)

	ctx.JSON(http.StatusOK, gin.H{
		"data": updatedGenreDTO,
	})
}

func (h *Handler) AddGenreToAnalysis(ctx *gin.Context) {
	userID := h.getCurrentUserID()
	genreID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный ID жанра"))
		return
	}

	err = h.Repository.AddGenreToAnalysis(userID, genreID)
	if err != nil {
		if err.Error() == "жанр уже добавлен в заявку" {
			h.errorHandler(ctx, http.StatusConflict, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.Status(http.StatusNoContent)
}
