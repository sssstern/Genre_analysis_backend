package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
// @Success 204 "Успешное начало асинхронного анализа / Отклонение"
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

	if action == "complete" {
		// 1. Меняем статус в БД на 'на анализе' и получаем DTO (требуется для получения TextToAnalyse для Log/Context)
		// Анализ DTO здесь используется для проверки, что заявка существует, и записи ModeratorID.
		analysisDTO, err := h.Repository.ChangeStatusToProcessing(uint(id), moderatorID)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
			return
		}

		// 2. Асинхронный вызов Django в отдельной горутине
		go func() {
			// 🔥 Вызываем только с ID, TextToAnalyse больше не нужен
			if err := h.CallDjangoAnalysisService(analysisDTO.AnalysisRequestID); err != nil {
				logrus.Errorf("Ошибка асинхронного вызова Django для заявки %d: %v", analysisDTO.AnalysisRequestID, err)
				h.Repository.HandleAnalysisFailure(uint(analysisDTO.AnalysisRequestID), "Ошибка вызова Django: "+err.Error())
			}
		}()

		// ВОЗВРАЩАЕМ 204 No Content
		ctx.Status(http.StatusNoContent)
		return
	}

	// ... (логика reject остается)
	if action == "reject" {
		analysisDTO, err := h.Repository.ProcessAnalysisRequest(uint(id), moderatorID, action)
		if err != nil {
			h.errorHandler(ctx, http.StatusBadRequest, err)
			return
		}
		h.successResponse(ctx, analysisDTO)
		return
	}

	h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("неизвестное действие: %s", action))
}

type CallDjangoRequest struct {
	AnalysisRequestID int    `json:"analysis_request_id"`
	TextToAnalyse     string `json:"text_to_analyse"`
	SecretKey         string `json:"secret_key"` // Для псевдо-авторизации
}

type AnalysisUpdateFromDjango struct {
	AnalysisRequestID int    `json:"analysis_request_id"`
	SecretKey         string `json:"secret_key"` // Для псевдо-авторизации
	// Используем именованную структуру
	AnalysisGenreData []ds.GenreUpdateData `json:"analysis_genre_data"`
}

// UpdateAnalysisResult
// @Summary Обработка результатов анализа от Django (Callback)
// @Description Внутренний маршрут. Не вызывается пользователем.
// @Tags Внутренний
// @Accept json
// @Produce json
// @Param data body AnalysisUpdateFromDjango true "Результаты анализа"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} handler.ErrorResponse "Неверный формат данных"
// @Failure 403 {object} handler.ErrorResponse "Неверный секретный ключ"
// @Failure 500 {object} handler.ErrorResponse "Ошибка базы данных"
// @Router /internal/update-analysis [post]
func (h *Handler) UpdateAnalysisResult(ctx *gin.Context) {
	var updateDTO AnalysisUpdateFromDjango

	if err := ctx.BindJSON(&updateDTO); err != nil {
		// 🔥 ЛОГИРОВАНИЕ: Выводим ошибку, если JSON пришел неверный
		logrus.Errorf("Callback JSON parsing error: %v", err)
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// 🔥 ЛОГИРОВАНИЕ: Выводим, что мы получили (для проверки данных)
	logrus.Infof("Callback received. RequestID: %d, Data: %+v", updateDTO.AnalysisRequestID, updateDTO.AnalysisGenreData)

	// Проверка секретного ключа для аутентификации callback'а (ИСПОЛЬЗУЕМ НОВЫЙ КЛЮЧ)
	if updateDTO.SecretKey != h.CallbackSecretKey {
		h.errorHandler(ctx, http.StatusForbidden, errors.New("неверный секретный ключ"))
		return
	}

	if err := h.Repository.UpdateAnalysisResults(updateDTO.AnalysisRequestID, updateDTO.AnalysisGenreData); err != nil {
		logrus.Errorf("Ошибка при обновлении результатов анализа для заявки %d: %v", updateDTO.AnalysisRequestID, err)
		h.errorHandler(ctx, http.StatusInternalServerError, errors.New("ошибка обновления результатов в базе данных"))
		return
	}
	ctx.Status(http.StatusNoContent)
	//ctx.JSON(http.StatusOK, gin.H{"message": "Результаты успешно приняты и обработаны"})
}

func (h *Handler) CallDjangoAnalysisService(analysisID int) error {
	// 🔥 УБРАЛИ callbackURL И secret_key ИЗ ТЕЛА ЗАПРОСА
	requestBody := map[string]interface{}{
		"analysis_request_id": analysisID,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга JSON: %w", err)
	}

	req, err := http.NewRequest("POST", h.DjangoServiceURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса к Django: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения запроса к Django: %w", err)
	}
	defer resp.Body.Close()

	// 🔥 Логика обработки 204 No Content остается
	if resp.StatusCode == http.StatusNoContent {
		logrus.Infof("Django service returned successful status 204 No Content for request %d", analysisID)
		return nil // Успешный запуск асинхронного процесса
	}

	// Обработка других успешных статусов (например, 200 OK)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		logrus.Infof("Django service returned successful status %d for request %d", resp.StatusCode, analysisID)
		return nil
	}

	// Обработка ошибок (4xx, 5xx)
	bodyBytes, _ := io.ReadAll(resp.Body)
	bodyStr := string(bodyBytes)
	if len(bodyStr) == 0 {
		bodyStr = "(Тело ответа пустое)"
	}

	logrus.Errorf("Django service returned error status %d: %s", resp.StatusCode, bodyStr)
	return fmt.Errorf("django service вернул ошибку: %s. Тело: %s", resp.Status, bodyStr)
}
