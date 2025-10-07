package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"lab3/internal/app/ds"

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

func (h *Handler) FormAnalysisRequest(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err = h.Repository.FormAnalysisRequest(uint(id))
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
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

	moderatorID := h.getCurrentModeratorID()
	analysisDTO, err := h.Repository.ProcessAnalysisRequest(uint(id), moderatorID, action)
	if err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	h.successResponse(ctx, analysisDTO)
}
