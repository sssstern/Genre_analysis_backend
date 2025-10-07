// repository/analysisrequest.go
package repository

import (
	"database/sql"
	"fmt"
	"lab3/internal/app/ds"
	"lab3/internal/app/service"
	"time"
)

func (r *Repository) UpdateAnalysisRequest(id uint, analysisUpdates ds.UpdateAnalysisRequestDTO) error {
	var analysis ds.AnalysisRequest
	err := r.db.Where("analysis_request_id = ? AND analysis_request_status = 'черновик'", id).First(&analysis).Error
	if err != nil {
		return err
	}

	if analysisUpdates.TextToAnalyse != "" {
		analysis.TextToAnalyse = analysisUpdates.TextToAnalyse
	}

	return r.db.Save(&analysis).Error
}

func (r *Repository) GetCurrentAnalysis(userID int) (*ds.AnalysisRequest, error) {
	// Этот метод используется внутренне, возвращает GORM-модель для логики. Если нужно, можно адаптировать, но для совместимости оставляем.
	var analysis ds.AnalysisRequest
	err := r.db.Where("creator_id = ? AND analysis_request_status = 'черновик'", userID).
		Preload("Genres").
		Preload("Genres.Genre").
		First(&analysis).Error

	if err != nil {
		if err.Error() == "record not found" {
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

func (r *Repository) GetCurrentAnalysisInfo(userID int) (int, int, error) {
	analysis, err := r.GetCurrentAnalysis(userID)
	if err != nil {
		return 0, 0, err
	}

	count := r.GetAnalysisCount(userID)
	return int(analysis.AnalysisRequestID), count, nil
}

func (r *Repository) GetAnalysisCount(userID int) int {
	analysis, err := r.GetCurrentAnalysis(userID)
	if err != nil || analysis == nil || analysis.AnalysisRequestID == 0 {
		return 0
	}

	var count int64
	err = r.db.Model(&ds.AnalysisGenre{}).
		Where("analysis_request_id = ?", analysis.AnalysisRequestID).
		Count(&count).Error
	if err != nil {
		return 0
	}

	return int(count)
}

func (r *Repository) AddGenreToAnalysis(userID, genreID int, comment string, probability int) error {
	analysis, err := r.GetCurrentAnalysis(userID)
	if err != nil {
		return err
	}

	var count int64
	err = r.db.Model(&ds.AnalysisGenre{}).
		Where("analysis_request_id = ? AND genre_id = ?", analysis.AnalysisRequestID, genreID).
		Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return fmt.Errorf("жанр уже добавлен в заявку")
	}

	item := ds.AnalysisGenre{
		AnalysisRequestID:  analysis.AnalysisRequestID,
		GenreID:            genreID,
		CommentToRequest:   comment,
		ProbabilityPercent: probability,
	}
	return r.db.Create(&item).Error
}

func (r *Repository) DeleteAnalysisRequest(analysisID uint) error {
	return r.db.Exec("UPDATE analysis_requests SET analysis_request_status = 'удалён' WHERE analysis_request_id = ?", analysisID).Error
}

func (r *Repository) GetAnalysisRequestByID(analysisID int) (*ds.AnalysisRequestDTO, error) {
	var analysis ds.AnalysisRequest
	err := r.db.Where("analysis_request_id = ? AND analysis_request_status != 'удалён'", analysisID).
		Preload("Genres").
		Preload("Genres.Genre").
		Preload("Creator").
		Preload("Moderator").
		First(&analysis).Error
	if err != nil {
		return nil, err
	}

	dto := &ds.AnalysisRequestDTO{
		AnalysisRequestID:     analysis.AnalysisRequestID,
		AnalysisRequestStatus: analysis.AnalysisRequestStatus,
		CreatedAt:             analysis.CreatedAt,
		CreatorLogin:          analysis.Creator.Login,
		TextToAnalyse:         analysis.TextToAnalyse,
	}
	if analysis.FormedAt.Valid {
		dto.FormedAt = &analysis.FormedAt.Time
	}
	if analysis.CompletedAt.Valid {
		dto.CompletedAt = &analysis.CompletedAt.Time
	}
	if analysis.ModeratorID.Valid {
		moderatorLogin := analysis.Moderator.Login
		dto.ModeratorLogin = &moderatorLogin
	}
	for _, ag := range analysis.Genres {
		dto.Genres = append(dto.Genres, ds.AnalysisGenreDTO{
			GenreID:            ag.GenreID,
			GenreName:          ag.Genre.GenreName,
			GenreImageURL:      ag.Genre.GenreImageURL,
			CommentToRequest:   ag.CommentToRequest,
			ProbabilityPercent: ag.ProbabilityPercent,
		})
	}
	return dto, nil
}

func (r *Repository) GetAnalysisRequests(status string, startDate, endDate time.Time) ([]ds.AnalysisRequestDTO, error) {
	// Логика поиска остается, но возвращаем []ds.AnalysisRequestDTO
	var analyses []ds.AnalysisRequest
	query := r.db.Where("analysis_request_status != 'удалён' AND analysis_request_status != 'черновик'")
	if status != "" {
		query = query.Where("analysis_request_status = ?", status)
	}
	if !startDate.IsZero() {
		query = query.Where("formed_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("formed_at <= ?", endDate)
	}
	err := query.Preload("Creator").Preload("Moderator").Preload("Genres.Genre").Find(&analyses).Error
	if err != nil {
		return nil, err
	}

	dtos := make([]ds.AnalysisRequestDTO, len(analyses))
	for i, a := range analyses {
		dto := ds.AnalysisRequestDTO{
			AnalysisRequestID:     a.AnalysisRequestID,
			AnalysisRequestStatus: a.AnalysisRequestStatus,
			CreatedAt:             a.CreatedAt,
			CreatorLogin:          a.Creator.Login,
			TextToAnalyse:         a.TextToAnalyse,
		}
		if a.FormedAt.Valid {
			dto.FormedAt = &a.FormedAt.Time
		}
		if a.CompletedAt.Valid {
			dto.CompletedAt = &a.CompletedAt.Time
		}
		if a.ModeratorID.Valid {
			moderatorLogin := a.Moderator.Login
			dto.ModeratorLogin = &moderatorLogin
		}
		for _, ag := range a.Genres {
			dto.Genres = append(dto.Genres, ds.AnalysisGenreDTO{
				GenreID:            ag.GenreID,
				GenreName:          ag.Genre.GenreName,
				GenreImageURL:      ag.Genre.GenreImageURL,
				CommentToRequest:   ag.CommentToRequest,
				ProbabilityPercent: ag.ProbabilityPercent,
			})
		}
		dtos[i] = dto
	}
	return dtos, nil
}

func (r *Repository) FormAnalysisRequestWithValidation(id uint) error {
	var analysis ds.AnalysisRequest
	err := r.db.Where("analysis_request_id = ? AND analysis_request_status = 'черновик'", id).First(&analysis).Error
	if err != nil {
		return err
	}

	if analysis.TextToAnalyse == "" {
		return fmt.Errorf("текст для анализа не может быть пустым")
	}

	var count int64
	err = r.db.Model(&ds.AnalysisGenre{}).Where("analysis_request_id = ?", id).Count(&count).Error
	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf("нельзя сформировать заявку без жанров")
	}

	analysis.AnalysisRequestStatus = "сформирован"
	analysis.FormedAt = sql.NullTime{Time: time.Now(), Valid: true}

	return r.db.Save(&analysis).Error
}

func (r *Repository) ProcessAnalysisRequest(id uint, moderatorID int, action string) (*ds.AnalysisRequestDTO, error) {
	var analysis ds.AnalysisRequest

	err := r.db.
		Where("analysis_request_id = ?", id).
		Preload("Creator").
		Preload("Moderator").
		Preload("Genres").
		Preload("Genres.Genre").
		First(&analysis).Error

	if err != nil {
		return nil, err
	}

	if analysis.AnalysisRequestStatus != "сформирован" {
		return nil, fmt.Errorf("заявка не может быть обработана, так как ее текущий статус: %s. Требуется статус 'сформирован'", analysis.AnalysisRequestStatus)
	}

	analysis.ModeratorID = sql.NullInt64{Int64: int64(moderatorID), Valid: true}

	if action == "complete" {
		textToAnalyse := analysis.TextToAnalyse

		for i := range analysis.Genres {
			ag := &analysis.Genres[i]

			if ag.Genre.GenreKeywords == "" {
				return nil, fmt.Errorf("не удалось найти ключевые слова для жанра ID %d", ag.GenreID)
			}

			keywords := ag.Genre.GenreKeywords
			probability := service.CalculateGenreProbability(textToAnalyse, keywords)
			ag.ProbabilityPercent = probability

			if err := r.db.Save(ag).Error; err != nil {
				return nil, fmt.Errorf("ошибка при сохранении вероятности жанра %d: %w", ag.GenreID, err)
			}
		}

		analysis.AnalysisRequestStatus = "завершён"
		analysis.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true}

	} else if action == "reject" {
		analysis.AnalysisRequestStatus = "отклонен"
	} else {
		return nil, fmt.Errorf("недопустимое действие: %s. Допустимо 'complete' или 'reject'", action)
	}

	err = r.db.Save(&analysis).Error
	if err != nil {
		return nil, err
	}

	dto := ds.AnalysisRequestDTO{
		AnalysisRequestID:     analysis.AnalysisRequestID,
		AnalysisRequestStatus: analysis.AnalysisRequestStatus,
		CreatedAt:             analysis.CreatedAt,
		CreatorLogin:          analysis.Creator.Login,
		TextToAnalyse:         analysis.TextToAnalyse,
	}

	if analysis.FormedAt.Valid {
		dto.FormedAt = &analysis.FormedAt.Time
	}
	if analysis.CompletedAt.Valid {
		dto.CompletedAt = &analysis.CompletedAt.Time
	}

	if analysis.ModeratorID.Valid {
		dto.ModeratorLogin = &analysis.Moderator.Login
	}

	for _, ag := range analysis.Genres {
		dto.Genres = append(dto.Genres, ds.AnalysisGenreDTO{
			GenreID:            ag.GenreID,
			GenreName:          ag.Genre.GenreName,
			GenreImageURL:      ag.Genre.GenreImageURL,
			CommentToRequest:   ag.CommentToRequest,
			ProbabilityPercent: ag.ProbabilityPercent,
		})
	}

	return &dto, nil
}

// Другие методы (UpdateAnalysisGenre, RemoveGenreFromAnalysis и т.д.) остаются, так как не возвращают DTO напрямую, но могут быть адаптированы если нужно.
func (r *Repository) UpdateAnalysisGenre(userID, genreID int, comment string, probability int) error {
	// Логика обновления m-m, без DTO возврата (success only)
	analysis, err := r.GetCurrentAnalysis(userID)
	if err != nil {
		return err
	}

	var ag ds.AnalysisGenre
	err = r.db.Where("analysis_request_id = ? AND genre_id = ?", analysis.AnalysisRequestID, genreID).First(&ag).Error
	if err != nil {
		return err
	}

	ag.CommentToRequest = comment
	ag.ProbabilityPercent = probability

	return r.db.Save(&ag).Error
}

func (r *Repository) RemoveGenreFromAnalysis(userID, genreID int) error {
	analysis, err := r.GetCurrentAnalysis(userID)
	if err != nil {
		return err
	}

	return r.db.Where("analysis_request_id = ? AND genre_id = ?", analysis.AnalysisRequestID, genreID).Delete(&ds.AnalysisGenre{}).Error
}
