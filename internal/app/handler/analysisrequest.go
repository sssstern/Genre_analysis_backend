package handler

import (
	"math/rand"
	"net/http"
	"strconv"

	"lab2/internal/app/helper"

	"github.com/gin-gonic/gin"
)

func (h *Handler) AddGenreToAnalysis(ctx *gin.Context) {
	userID := 1 // Хардкор ID пользователя

	// Получаем ID жанра из формы
	genreIDStr := ctx.PostForm("genre_id")
	genreID, err := strconv.Atoi(genreIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	// Добавляем жанр в заявку
	err = h.Repository.AddGenreToAnalysis(userID, genreID, helper.GetRandomPhrase(), rand.Intn(100))
	if err != nil {
		if err.Error() == "жанр уже добавлен в заявку" {
			ctx.Redirect(http.StatusFound, "/genres?error=genre_already_added")
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	// Перенаправляем обратно на страницу жанров
	ctx.Redirect(http.StatusFound, "/genres")
}

func (h *Handler) GetAnalysis(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	analysis, err := h.Repository.GetAnalysisByID(uint(id))
	if err != nil {
		// Если заявка не найдена или удалена - возвращаем 404
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	if len(analysis.Genres) == 0 {
		ctx.Redirect(http.StatusFound, "/genres")
		return
	}

	ctx.HTML(http.StatusOK, "analysis.html", gin.H{
		"analysis": analysis,
	})
}

func (h *Handler) DeleteAnalysis(ctx *gin.Context) {
	analysisIDStr := ctx.PostForm("analysis_request_id")
	analysisID, err := strconv.Atoi(analysisIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.DeleteAnalysis(uint(analysisID))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Redirect(http.StatusFound, "/genres")
}
