package handler

import (
	"lab3/internal/app/ds"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func (h *Handler) UpdateAnalysisGenre(ctx *gin.Context) {
	userID := h.getCurrentUserID()

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

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func (h *Handler) RemoveGenreFromAnalysis(ctx *gin.Context) {
	userID := h.getCurrentUserID()

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
