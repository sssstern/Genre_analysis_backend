package handler

import (
	"net/http"
	"strconv"

	"lab2/internal/app/ds"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetGenres(ctx *gin.Context) {
	userID := int(1) // Захардкоженный ID пользователя

	var genres []ds.Genre
	var err error

	searchQuery := ctx.Query("searchbygenrename")
	if searchQuery == "" {
		genres, err = h.Repository.GetGenres()
	} else {
		genres, err = h.Repository.GetGenresByTitle(searchQuery)
	}

	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	analysisCount := h.Repository.GetAnalysisCount(userID)
	currentAnalysis, _ := h.Repository.GetCurrentAnalysis(userID)

	ctx.HTML(http.StatusOK, "genres.html", gin.H{
		"genres":            genres,
		"searchbygenrename": searchQuery,
		"analysisCount":     analysisCount,
		"currentAnalysisID": currentAnalysis.AnalysisRequestID,
	})
}

func (h *Handler) GetGenre(ctx *gin.Context) {
	userID := int(1)

	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	genre, err := h.Repository.GetGenre(id)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	analysisCount := h.Repository.GetAnalysisCount(userID)
	currentAnalysis, _ := h.Repository.GetCurrentAnalysis(userID)

	ctx.HTML(http.StatusOK, "genre.html", gin.H{
		"genre":             genre,
		"analysisCount":     analysisCount,
		"currentAnalysisID": currentAnalysis.AnalysisRequestID,
	})
}
