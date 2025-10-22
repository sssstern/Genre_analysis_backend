package handler

import (
	"lab4/internal/app/ds"
	"lab4/internal/app/middleware"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// UpdateAnalysisGenre
// @Summary Обновить жанр в заявке
// @Description Обновляет комментарий и процент вероятности для жанра в текущем черновике заявки.
// @Tags Домен м-м
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param id path int true "ID жанра"
// @Param request body ds.UpdateGenreRequestDTO true "Комментарий и процент вероятности"
// @Success 204 "Успешное обновление"
// @Failure 400 {object} handler.ErrorResponse "Неверный формат ID или данных"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован"
// @Failure 500 {object} handler.ErrorResponse "Ошибка сервера"
// @Router /analysis-genres/{id} [put]
func (h *Handler) UpdateAnalysisGenre(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)

	genreIDStr := ctx.Param("id")
	genreID, err := strconv.Atoi(genreIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req ds.UpdateGenreRequestDTO
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}
	err = h.Repository.UpdateAnalysisGenre(userID, genreID, req.CommentToRequest, req.ProbabilityPercent)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// RemoveGenreFromAnalysis
// @Summary Удалить жанр из заявки
// @Description Удаляет жанр из текущего черновика заявки.
// @Tags Домен м-м
// @Produce json
// @Param id path int true "ID жанра для удаления"
// @Success 204 "Успешное удаление"
// @Failure 400 {object} handler.ErrorResponse "Неверный формат ID"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован"
// @Failure 500 {object} handler.ErrorResponse "Ошибка сервера"
// @Router /analysis-genres/{id} [delete]
func (h *Handler) RemoveGenreFromAnalysis(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)

	genreIDStr := ctx.Param("id")
	genreID, err := strconv.Atoi(genreIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.RemoveGenreFromAnalysis(userID, genreID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Genre removed from cart successfully",
	})
}
