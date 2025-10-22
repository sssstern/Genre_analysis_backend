// handler/genre.go
package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"lab4/internal/app/ds"
	"lab4/internal/app/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// GetGenres
// @Summary Получить список жанров
// @Description Возвращает список всех существующих жанров. Доступен публично.
// @Tags Домен жанров
// @Produce json
// @Param searchbygenrename query string false "Поиск по названию жанра (частичное совпадение)"
// @Success 200 {object} ds.GenreDTO "Список жанров"
// @Failure 500 {object} handler.ErrorResponse "Ошибка сервера"
// @Router /genres [get]
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

// GetGenre
// @Summary Получить жанр по ID
// @Description Возвращает информацию о конкретном жанре. Доступен публично.
// @Tags Домен жанров
// @Produce json
// @Param id path int true "ID жанра"
// @Success 200 {object} ds.GenreDTO "Информация о жанре"
// @Failure 400 {object} handler.ErrorResponse "Неверный формат ID"
// @Failure 404 {object} handler.ErrorResponse "Жанр не найден"
// @Router /genres/{id} [get]
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

// CreateGenre
// @Summary Создать новый жанр
// @Description Создает новый жанр. Требуются права **Модератора**.
// @Tags Домен жанров
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param request body ds.UpdateGenreDTO true "Название и ключевые слова жанра"
// @Success 201 {object} ds.GenreDTO "Успешное создание жанра"
// @Failure 400 {object} handler.ErrorResponse "Неверный формат данных"
// @Failure 403 {object} handler.ErrorResponse "Доступ запрещен (не модератор)"
// @Failure 500 {object} handler.ErrorResponse "Ошибка сервера"
// @Router /genres [post]
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

// UpdateGenre
// @Summary Обновить жанр
// @Description Обновляет название и/или ключевые слова жанра по ID. Требуются права **Модератора**.
// @Tags Домен жанров
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param id path int true "ID жанра"
// @Param request body ds.UpdateGenreDTO true "Новые название и ключевые слова жанра"
// @Success 200 {object} ds.GenreDTO "Успешное обновление"
// @Failure 400 {object} handler.ErrorResponse  "Неверный формат ID или данных"
// @Failure 403 {object} handler.ErrorResponse  "Доступ запрещен (не модератор)"
// @Failure 404 {object} handler.ErrorResponse  "Жанр не найден"
// @Router /genres/{id} [put]
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

// DeleteGenre
// @Summary Удалить жанр
// @Description Устанавливает флаг is_deleted = true для жанра. Требуются права **Модератора**.
// @Tags Домен жанров
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param id path int true "ID жанра"
// @Success 204 "Успешное удаление"
// @Failure 400 {object} handler.ErrorResponse"Неверный формат ID"
// @Failure 403 {object} handler.ErrorResponse "Доступ запрещен (не модератор)"
// @Failure 500 {object} handler.ErrorResponse "Ошибка сервера"
// @Router /genres/{id} [delete]
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

// UploadGenreImage
// @Summary Загрузить изображение жанра
// @Description Загружает и обновляет изображение для жанра по ID. Требуются права **Модератора**.
// @Tags Домен жанров
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param id path int true "ID жанра"
// @Param file formData file true "Файл изображения"
// @Success 200 {object} ds.GenreDTO "Успешная загрузка, возвращает обновленный жанр"
// @Failure 400 {object} handler.ErrorResponse "Ошибка загрузки/формата файла"
// @Failure 403 {object} handler.ErrorResponse "Доступ запрещен (не модератор)"
// @Failure 500 {object} handler.ErrorResponse "Ошибка Minio/сервера"
// @Router /genres/{id}/image [post]
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

// AddGenreToAnalysis (в handler/genre.go, но относится к Заявкам)
// @Summary Добавить жанр в черновик заявки
// @Description Добавляет жанр в текущую черновую заявку пользователя. Требуется **Авторизация**.
// @Tags Домен жанров
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param id path int true "ID жанра для добавления"
// @Success 204 "Успешное добавление"
// @Failure 400 {object} handler.ErrorResponse"Неверный формат ID"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован"
// @Failure 500 {object} handler.ErrorResponse "Ошибка сервера (например, жанр уже добавлен)"
// @Router /genres/add-to-analysis/{id} [post]
func (h *Handler) AddGenreToAnalysis(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)
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
