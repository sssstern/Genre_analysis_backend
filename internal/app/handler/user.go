// handler/user.go
package handler

import (
	"fmt"
	"net/http"

	"lab3/internal/app/ds"

	"github.com/gin-gonic/gin"
)

func (h *Handler) RegisterUser(ctx *gin.Context) {
	var input ds.RegisterUserRequestDTO
	if err := ctx.BindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if input.Login == "" || input.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("логин и пароль обязательны"))
		return
	}

	err := h.Repository.RegisterUser(input)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"status": "success",
	})
}

func (h *Handler) LoginUser(ctx *gin.Context) {
	var input struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err := ctx.BindJSON(&input); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	if input.Login == "" || input.Password == "" {
		h.errorHandler(ctx, http.StatusBadRequest, fmt.Errorf("логин и пароль обязательны"))
		return
	}

	userDTO, err := h.Repository.LoginUser(input.Login, input.Password)
	if err != nil {
		h.errorHandler(ctx, http.StatusUnauthorized, err)
		return
	}

	h.successResponse(ctx, userDTO)
}

func (h *Handler) GetProfile(ctx *gin.Context) {
	userID := h.getCurrentUserID()

	userDTO, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	h.successResponse(ctx, userDTO)
}

func (h *Handler) UpdateProfile(ctx *gin.Context) {
	userID := h.getCurrentUserID()

	var userUpdates ds.UpdateUserRequestDTO
	if err := ctx.BindJSON(&userUpdates); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	updatedUserDTO, err := h.Repository.UpdateUser(userID, userUpdates)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	h.successResponse(ctx, updatedUserDTO)
}

func (h *Handler) LogoutUser(ctx *gin.Context) {
	userID := h.getCurrentUserID()

	err := h.Repository.LogoutUser(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Успешный выход из системы",
	})
}
