// handler/analysisrequest.go
package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"lab3/internal/app/ds"
	"lab3/internal/app/service"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetCurrentAnalysis(ctx *gin.Context) {
	userID := h.getCurrentUserID()

	currentAnalysisID, count, err := h.Repository.GetCurrentAnalysisInfo(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	response := gin.H{
		"analysis_request_id":     currentAnalysisID,
		"genres_in_request_count": count,
		"analysis_request_img":    "/img/RequestIcon.png",
	}

	h.successResponse(ctx, response)
}

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

	err = h.Repository.UpdateAnalysisRequest(uint(id), analysisUpdates)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

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

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func (h *Handler) AddGenreToAnalysis(ctx *gin.Context) {
	userID := h.getCurrentUserID()

	genreID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "Неверный ID жанра",
		})
		return
	}

	err = h.Repository.AddGenreToAnalysis(userID, genreID, service.GetRandomPhrase(), rand.Intn(100))
	if err != nil {
		if err.Error() == "жанр уже добавлен в заявку" {
			h.errorHandler(ctx, http.StatusConflict, err)
		} else {
			h.errorHandler(ctx, http.StatusInternalServerError, err)
		}
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
	})
}

type UpdateCartGenreRequest struct {
	CommentToRequest   string `json:"comment_to_request"`
	ProbabilityPercent int    `json:"probability_percent"`
}

func (h *Handler) UpdateAnalysisGenre(ctx *gin.Context) {
	userID := h.getCurrentUserID()

	genreIDStr := ctx.Param("id")
	genreID, err := strconv.Atoi(genreIDStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req UpdateCartGenreRequest
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

func (h *Handler) GetAnalysisRequests(ctx *gin.Context) {
	status := ctx.Query("status")
	startDateStr := ctx.Query("start_date")
	endDateStr := ctx.Query("end_date")

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
	requests, err := h.Repository.GetAnalysisRequests(status, startDate, endDate)

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

func (h *Handler) FormAnalysisRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.FormAnalysisRequestWithValidation(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

type ProcessAnalysisRequestRequest struct {
	Action string `json:"action" binding:"required"` // "complete" или "reject"
}

func (h *Handler) ProcessAnalysisRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	var req ProcessAnalysisRequestRequest
	if err := ctx.BindJSON(&req); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	moderatorID := 2 // хардкор модератора
	analysisDTO, err := h.Repository.ProcessAnalysisRequest(uint(id), moderatorID, req.Action)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	h.successResponse(ctx, analysisDTO)
}
