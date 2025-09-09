package handler

import (
	"lab1/internal/app/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetGenres(ctx *gin.Context) {
	var genres []repository.Genre
	var err error

	searchQuery := ctx.Query("query") // получаем значение из поля поиска
	if searchQuery == "" {            // если поле поиска пусто, то получаем все жанры
		genres, err = h.Repository.GetGenres()
		if err != nil {
			logrus.Error("Ошибка получения жанров:", err)
			return
		}
	} else {
		genres, err = h.Repository.GetGenresByTitle(searchQuery) // ищем жанры по запросу
		if err != nil {
			logrus.Error("Ошибка поиска жанров:", err)
			return
		}
	}

	analysisCount := h.Repository.GetAnalysisCount()
	currentAnalysis := h.Repository.GetCurrentAnalysis()

	ctx.HTML(http.StatusOK, "genres.html", gin.H{
		"time":              time.Now().Format("15:04:05"),
		"genres":            genres,
		"query":             searchQuery,                            // передаем запрос обратно в форму
		"AnalysisCount":     analysisCount,                          // Количество услуг
		"currentAnalysisID": currentAnalysis.GenreAnalysisRequestID, // ID заявки
	})

}

func (h *Handler) GetGenre(ctx *gin.Context) {
	idStr := ctx.Param("id") // получаем id жанра из URL (/genre/:id)
	id, err := strconv.Atoi(idStr)
	if err != nil {
		logrus.Error("Неверный ID жанра:", err)
		return
	}

	genre, err := h.Repository.GetGenre(id)
	if err != nil {
		logrus.Error("Ошибка получения жанра:", err)
	}

	analysisCount := h.Repository.GetAnalysisCount()
	currentAnalysis := h.Repository.GetCurrentAnalysis()

	ctx.HTML(http.StatusOK, "genre.html", gin.H{
		"genre":             genre,
		"time":              time.Now().Format("15:04:05"),
		"AnalysisCount":     analysisCount, // Количество услуг
		"currentAnalysisID": currentAnalysis.GenreAnalysisRequestID,
	})
}

func (h *Handler) GetAnalysis(ctx *gin.Context) {
	analysis := h.Repository.GetCurrentAnalysis()

	allGenres, err := h.Repository.GetGenres()
	if err != nil {
		logrus.Error("Ошибка получения жанров:", err)
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Не удалось загрузить список жанров",
		})
		return
	}

	genreMap := make(map[int]repository.Genre)
	for _, genre := range allGenres {
		genreMap[genre.GenreID] = genre
	}

	ctx.HTML(http.StatusOK, "analysis.html", gin.H{
		"analysis": analysis,
		"genreMap": genreMap,
		"time":     time.Now().Format("15:04:05"),
	})
}
