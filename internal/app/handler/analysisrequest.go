package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"lab4/internal/app/ds"

	"lab4/internal/app/middleware"

	"context" // Нужно для Redis-клиента
	"errors"  // Нужно для проверки ошибки redis.Nil

	"github.com/gin-gonic/gin"

	"github.com/redis/go-redis/v9" // Если вы используете go-redis/v9
	"github.com/sirupsen/logrus"

	"lab4/internal/app/service" // Для ExtractToken и ParseJWT
)

// GetCurrentAnalysis
// @Summary Получить информацию о черновике
// @Description Возвращает ID текущей черновой заявки и количество жанров в ней.
// @Tags Домен заявки на анализ текста
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Success 200 {object} object{analysis_request_id=int,genres_in_request_count=int} "Информация о черновике"
// @Failure 500 {object} handler.ErrorResponse "Ошибка сервера"
// @Router /text-analysis-request/icon [get]
func (h *Handler) GetCurrentAnalysis(ctx *gin.Context) {
	tokenString := service.ExtractToken(ctx)

	var userID int = 0

	if tokenString != "" {
		val, redisErr := h.RedisClient.Get(context.Background(), tokenString).Result()

		if (redisErr == nil && val == "blacklist") || (redisErr != nil && !errors.Is(redisErr, redis.Nil)) {
			logrus.Warnf("Token is blacklisted or Redis error (Guest status): %v", redisErr)
		} else {
			claims, parseErr := service.ParseJWT(tokenString, h.SecretKey)

			if parseErr == nil {
				userID = claims.UserID
			}
		}
	}

	if userID == 0 {
		ctx.JSON(http.StatusOK, gin.H{
			"analysis_request_id":     0,
			"genres_in_request_count": 0,
		})
		return
	}

	currentAnalysisID, count, err := h.Repository.GetCurrentAnalysisInfo(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	response := gin.H{
		"analysis_request_id":     currentAnalysisID,
		"genres_in_request_count": count,
	}

	h.successResponse(ctx, response)
}

// GetAnalysisRequests
// @Summary Получить список заявок
// @Description Для модератора - все заявки. Для создателя - только его заявки.
// @Tags Домен заявки на анализ текста
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param status query string false "Фильтр по статусу ('черновик', 'сформирован', 'завершён', 'отклонён')"
// @Success 200 {array} []ds.AnalysisRequestDTO "Список заявок"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован"
// @Router /text-analysis-request [get]
func (h *Handler) GetAnalysisRequests(ctx *gin.Context) {
	status := ctx.Query("status")
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

	userID := middleware.GetUserID(ctx)
	userRole := middleware.GetRole(ctx)

	var startDate time.Time
	var endDate time.Time
	var err error

	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат start_date. Используйте YYYY-MM-DD"))
			return
		}
	}

	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неверный формат end_date. Используйте YYYY-MM-DD"))
			return
		}
	}
	requests, err := h.Repository.GetAnalysisRequests(userID, userRole, status, startDate, endDate)

	if err != nil {
		if err.Error() == "record not found" {
			h.successResponse(ctx, []ds.AnalysisRequestDTO{})
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successResponse(ctx, requests)
}

// GetAnalysisRequest
// @Summary Получить одну заявку по id
// @Tags Домен заявки на анализ текста
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param id path int true "ID заявки"
// @Success 200 {array} ds.AnalysisRequestDTO "Заявка"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован"
// @Router /text-analysis-request/{id} [get]
func (h *Handler) GetAnalysisRequestByID(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	requestDTO, err := h.Repository.GetAnalysisRequestByID(id)
	if err != nil {
		if err.Error() == "record not found" {
			h.errorHandler(ctx, http.StatusNotFound, fmt.Errorf("заявка с ID %d не найдена", id))
			return
		}
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successResponse(ctx, requestDTO)
}

// UpdateAnalysisRequest
// @Summary Обновить черновик
// @Description Обновляет поле 'TextToAnalyse' черновой заявки.
// @Tags Домен заявки на анализ текста
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param id path int true "ID заявки (черновика)"
// @Param request body ds.UpdateAnalysisRequestDTO true "Новый текст для анализа"
// @Success 200 {object} ds.AnalysisRequestDTO "Успешное обновление"
// @Failure 400 {object} handler.ErrorResponse"Неверный формат ID или данных"
// @Failure 401 {object} handler.ErrorResponse"Неавторизован"
// @Failure 500 {object} handler.ErrorResponse "Ошибка сервера/Не является черновиком"
// @Router /text-analysis-requests/{id} [put]
func (h *Handler) UpdateAnalysisRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var analysisUpdates ds.UpdateAnalysisRequestDTO
	if err := ctx.BindJSON(&analysisUpdates); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	_, err = h.Repository.UpdateAnalysisRequest(uint(id), analysisUpdates)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	updatedAnalysisDTO, err := h.Repository.GetAnalysisRequestByID(int(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("обновление успешно, но ошибка при получении данных для ответа: %w", err))
		return
	}

	h.successResponse(ctx, updatedAnalysisDTO)
}

// FormAnalysisRequest
// @Summary Отправить заявку на модерацию
// @Description Переводит статус черновика на 'на модерации'.
// @Tags Домен заявки на анализ текста
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param id path int true "ID заявки (черновика)"
// @Success 200 {object} ds.AnalysisRequestDTO "Успешная отправка"
// @Failure 400 {object} handler.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован"
// @Failure 500 {object} handler.ErrorResponse "Ошибка сервера/Не является черновиком"
// @Router /text-analysis-requests/{id}/form [put]
func (h *Handler) FormAnalysisRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	_, err = h.Repository.FormAnalysisRequest(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	updatedAnalysisDTO, err := h.Repository.GetAnalysisRequestByID(int(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("обновление успешно, но ошибка при получении данных для ответа: %w", err))
		return
	}

	h.successResponse(ctx, updatedAnalysisDTO)
}

// DeleteAnalysisRequest
// @Summary Удалить черновик заявки
// @Description Удаляет заявку (только черновик).
// @Tags Домен заявки на анализ текста
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param id path int true "ID заявки (черновика)"
// @Success 204 "Успешное удаление"
// @Failure 400 {object} handler.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован"
// @Failure 500 {object} handler.ErrorResponse "Ошибка сервера/Не является черновиком"
// @Router /text-analysis-requests/{id} [delete]
func (h *Handler) DeleteAnalysisRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeleteAnalysisRequest(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ProcessAnalysisRequest godoc
// @Summary Завершить или отклонить заявку (Только для Модератора)
// @Tags Домен заявки на анализ текста
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param id path int true "ID заявки"
// @Param action query string true "Действие ('complete' или 'reject')"
// @Success 200 {object} ds.AnalysisRequestDTO "Обновленная заявка"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован"
// @Failure 403 {object} handler.ErrorResponse "Доступ запрещен (не модератор)"
// @Failure 404 {object} handler.ErrorResponse "Заявка не найдена"
// @Router /text-analysis-requests/{id}/process [put]
func (h *Handler) ProcessAnalysisRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	action := ctx.Query("action")
	if action == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("параметр action обязателен"))
		return
	}

	moderatorID := middleware.GetUserID(ctx)
	analysisDTO, err := h.Repository.ProcessAnalysisRequest(uint(id), moderatorID, action)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	h.successResponse(ctx, analysisDTO)
}
