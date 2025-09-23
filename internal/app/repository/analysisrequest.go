package repository

import (
	"fmt"
	"lab2/internal/app/ds"
	"time"
)

func (r *Repository) GetCurrentAnalysis(userID int) (*ds.AnalysisRequest, error) {
	var analysis ds.AnalysisRequest
	err := r.db.Where("creator_id = ? AND analysis_request_status = 'черновик'", userID).
		Preload("Genres").
		Preload("Genres.Genre").
		First(&analysis).Error

	if err != nil {
		if err.Error() == "record not found" {
			// Создаем новую заявку
			newAnalysis := &ds.AnalysisRequest{
				AnalysisRequestStatus: "черновик",
				TextToAnalyse:         "Ранним утром солнце медленно поднималось над горизонтом, окрашивая небо в нежные розовые тона. Вдалеке шумел лес, наполняя воздух свежестью и ароматом хвои. Маленький ручеёк извивался между камнями, неся свои воды к большой реке. Птицы начинали свой дневной концерт, наполняя лес мелодичными трелями. Этот уголок природы был настоящим оазисом спокойствия и гармонии среди шумного города.",
				CreatorID:             userID,
				CreatedAt:             time.Now(),
			}

			err = r.db.Create(newAnalysis).Error
			if err != nil {
				return nil, err
			}

			// Загружаем созданную заявку с отношениями
			err = r.db.Where("analysis_request_id = ?", newAnalysis.AnalysisRequestID).
				Preload("Genres").
				Preload("Genres.Genre").
				First(&analysis).Error
			if err != nil {
				return nil, err
			}

			return &analysis, nil
		}
		return nil, err
	}

	return &analysis, nil
}

func (r *Repository) GetAnalysisCount(userID int) int {
	analysis, err := r.GetCurrentAnalysis(userID)
	if err != nil || analysis == nil || analysis.AnalysisRequestID == 0 {
		return 0 // Возвращаем 0, если ошибка или нет данных
	}

	var count int64
	err = r.db.Model(&ds.AnalysisGenre{}).
		Where("analysis_request_id = ?", analysis.AnalysisRequestID).
		Count(&count).Error
	if err != nil {
		return 0 // Возвращаем 0, если ошибка при подсчёте
	}

	return int(count)
}

func (r *Repository) AddGenreToAnalysis(userID, genreID int, comment string, probability int) error {
	// Получаем текущую заявку или создаём новую, если её нет
	analysis, err := r.GetCurrentAnalysis(userID)
	if err != nil {
		return err
	}

	// Проверяем, не добавлен ли уже этот жанр в заявку
	var count int64
	err = r.db.Model(&ds.AnalysisGenre{}).
		Where("analysis_request_id = ? AND genre_id = ?", analysis.AnalysisRequestID, genreID).
		Count(&count).Error
	if err != nil {
		return err
	}

	// Если жанр уже добавлен, возвращаем ошибку
	if count > 0 {
		return fmt.Errorf("жанр уже добавлен в заявку")
	}

	// Добавляем жанр в заявку
	item := ds.AnalysisGenre{
		AnalysisRequestID:  analysis.AnalysisRequestID,
		GenreID:            genreID,
		CommentToRequest:   comment,
		ProbabilityPercent: probability,
	}
	return r.db.Create(&item).Error
}

func (r *Repository) DeleteAnalysis(analysisID uint) error {
	return r.db.Exec("UPDATE analysis_requests SET analysis_request_status = 'удалён' WHERE analysis_request_id = ?", analysisID).Error
}

func (r *Repository) GetAnalysisByID(analysisID uint) (*ds.AnalysisRequest, error) {
	var analysis ds.AnalysisRequest
	err := r.db.Where("analysis_request_id = ? AND analysis_request_status != 'удалён'", analysisID).
		Preload("Genres").
		Preload("Genres.Genre").
		First(&analysis).Error

	if err != nil {
		return nil, err
	}

	return &analysis, nil
}
