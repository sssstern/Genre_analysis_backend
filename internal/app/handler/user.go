// handler/user.go
package handler

import (
	"fmt"
	"net/http"
	"time"

	"lab4/internal/app/ds"
	"lab4/internal/app/middleware"
	"lab4/internal/app/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// @Summary Регистрация нового пользователя
// @Tags Домен пользователя
// @Accept json
// @Produce json
// @Param request body ds.ChangeUserDTO true "Данные для регистрации (логин и пароль)"
// @Success 204 "Успешная регистрация"
// @Failure 400 {object} handler.ErrorResponse  "Неверный формат данных"
// @Failure 500 {object} handler.ErrorResponse  "Ошибка сервера или пользователь с таким логином уже существует"
// @Router /user/register [post]
func (h *Handler) RegisterUser(ctx *gin.Context) {
	var input ds.ChangeUserDTO
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

	ctx.Status(http.StatusNoContent)
}

// @Summary Аутентификация пользователя
// @Tags Домен пользователя
// @Accept  json
// @Produce json
// @Param user body ds.ChangeUserDTO true "Данные для входа"
// @Success 200 {object} ds.AuthResponseDTO "Успешный вход"
// @Failure 400 {object} handler.ErrorResponse "Неверный запрос"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован"
// @Router /user/login [post]
func (h *Handler) LoginUser(ctx *gin.Context) {
	var input ds.ChangeUserDTO
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

	tokenString, expTime, err := service.GenerateJWT(userDTO.UserID, userDTO.Role, h.SecretKey, h.JWTDur)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка генерации токена: %w", err))
		return
	}

	ctx.SetCookie("session_token", tokenString, int(expTime.Unix()), "/", h.HostName, false, true)

	h.successResponse(ctx, ds.AuthResponseDTO{
		AccessToken: tokenString,
		TokenType:   "Bearer",
		ExpiresIn:   expTime.Unix(),
	})
}

// GetProfile
// @Summary Получить профиль
// @Tags Домен пользователя
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Success 200 {object} ds.UserDTO "Данные пользователя"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован"
// @Failure 404 {object} handler.ErrorResponse  "Пользователь не найден (редкий случай)"
// @Router /user/profile [get]
func (h *Handler) GetProfile(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)

	userDTO, err := h.Repository.GetUserByID(userID)
	if err != nil {
		h.errorHandler(ctx, http.StatusNotFound, err)
		return
	}

	h.successResponse(ctx, userDTO)
}

// UpdateProfile
// @Summary Обновить профиль
// @Tags Домен пользователя
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Param request body ds.ChangeUserDTO true "Новые данные (логин/пароль)"
// @Success 200 {object} ds.UserDTO "Успешное обновление"
// @Failure 400 {object} handler.ErrorResponse "Неверный формат данных"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован"
// @Failure 500 {object} handler.ErrorResponse "Ошибка сервера"
// @Router /user/profile [put]
func (h *Handler) UpdateProfile(ctx *gin.Context) {
	userID := middleware.GetUserID(ctx)

	var userUpdates ds.ChangeUserDTO
	if err := ctx.BindJSON(&userUpdates); err != nil {
		h.errorHandler(ctx, http.StatusBadRequest, err)
		return
	}

	err := h.Repository.UpdateUser(userID, userUpdates)
	if err != nil {
		h.errorHandler(ctx, http.StatusInternalServerError, err)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// LogoutUser
// @Summary Выход из системы
// @Tags Домен пользователя
// @Produce json
// @Security ApiKeyAuth
// @Security SessionCookie
// @Success 204 "Успешный выход"
// @Failure 401 {object} handler.ErrorResponse "Неавторизован (отсутствует токен)"
// @Failure 500 {object} handler.ErrorResponse "Ошибка Redis/сервера"
// @Router /user/logout [post]
func (h *Handler) LogoutUser(ctx *gin.Context) {
	tokenString := service.ExtractToken(ctx)
	if tokenString == "" {
		h.errorHandler(ctx, http.StatusUnauthorized, fmt.Errorf("отсутствует токен для выхода"))
		return
	}

	claims, err := service.ParseJWT(tokenString, h.SecretKey)
	if err != nil {
		ctx.Status(http.StatusNoContent)
		return
	}

	remainingDur := time.Until(claims.ExpiresAt.Time)

	if remainingDur > 0 {
		err := h.Repository.AddToBlacklist(ctx, tokenString, remainingDur)
		if err != nil {
			h.errorHandler(ctx, http.StatusInternalServerError, fmt.Errorf("ошибка при добавлении в блеклист: %w", err))
			return
		}
	} else {
		logrus.Warnf("Попытка выхода с просроченным токеном. TTL: %v", remainingDur)
	}

	ctx.SetCookie("session_token", "", -1, "/", h.HostName, false, true)

	ctx.JSON(http.StatusOK, gin.H{"message": "Выход выполнен успешно"})
}
