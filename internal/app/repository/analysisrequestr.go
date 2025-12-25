package repository

import (
	"database/sql"
	"fmt"
	"lab4/internal/app/ds"

	//"lab4/internal/app/service"
	"time"

	"github.com/sirupsen/logrus"
)

func (r *Repository) GetCurrentAnalysisInfo(userID int) (currentAnalysisID int, count int64, err error) {
	logrus.Infof("GetCurrentAnalysis called for userID: %d", userID)
	var analysis ds.AnalysisRequest

	err = r.db.Where("creator_id = ? AND analysis_request_status = 'черновик'", userID).First(&analysis).Error
	if err != nil {
		if err.Error() == "record not found" {

			return 0, 0, nil
		}
		return 0, 0, err
	}

	err = r.db.Model(&ds.AnalysisGenre{}).Where("analysis_request_id = ?", analysis.AnalysisRequestID).Count(&count).Error
	if err != nil {
		return 0, 0, err
	}

	return analysis.AnalysisRequestID, count, nil
}

func (r *Repository) GetCurrentAnalysis(userID int) (*ds.AnalysisRequest, error) {
	logrus.Infof("GetCurrentAnalysis called for userID: %d", userID)
	var analysis ds.AnalysisRequest
	err := r.db.Where("creator_id = ? AND analysis_request_status = 'черновик'", userID).
		Preload("Genres").
		Preload("Genres.Genre").
		First(&analysis).Error

	if err != nil {
		if err.Error() == "record not found" {
			newAnalysis := &ds.AnalysisRequest{
				AnalysisRequestStatus: "черновик",
				TextToAnalyse:         "",
				CreatorID:             userID,
				CreatedAt:             time.Now(),
			}

			if err := r.db.Create(newAnalysis).Error; err != nil {
				return nil, err
			}
			return newAnalysis, nil
		}
		return nil, err
	}

	return &analysis, nil
}

func (r *Repository) GetAnalysisRequests(userID int, role ds.UserRole, status string, startDate, endDate time.Time) ([]ds.AnalysisRequestDTO, error) {
	query := r.db.Model(&ds.AnalysisRequest{}).
		Preload("Creator").
		Preload("Moderator").
		Preload("Genres").
		Preload("Genres.Genre")

	if role == ds.RoleCreator && userID != 0 {
		query = query.Where("creator_id = ?", userID)
	}
	query = query.Where("analysis_request_status != ?", "удалён")

	if status != "" {
		query = query.Where("analysis_request_status = ?", status)
	}
	if !startDate.IsZero() {
		query = query.Where("formed_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("formed_at <= ?", endDate)
	}

	var analyses []ds.AnalysisRequest
	err := query.Find(&analyses).Error
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
		var completedCount int64
		// Считаем записи м-м, где ProbabilityPercent > 0
		r.db.Model(&ds.AnalysisGenre{}).
			Where("analysis_request_id = ? AND probability_percent > 0", a.AnalysisRequestID).
			Count(&completedCount)

		dto.GenresCompletedCount = int(completedCount)
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
			GenreKeywords:      ag.Genre.GenreKeywords,
			CommentToRequest:   ag.CommentToRequest,
			ProbabilityPercent: ag.ProbabilityPercent,
		})
	}
	return dto, nil
}

func (r *Repository) UpdateAnalysisRequest(id uint, analysisUpdates ds.UpdateAnalysisRequestDTO) (*ds.AnalysisRequest, error) {
	var analysis ds.AnalysisRequest

	err := r.db.Where("analysis_request_id = ? AND analysis_request_status = 'черновик'", id).First(&analysis).Error
	if err != nil {
		return nil, err
	}

	if analysisUpdates.TextToAnalyse != "" {
		analysis.TextToAnalyse = analysisUpdates.TextToAnalyse
	}

	err = r.db.Save(&analysis).Error
	if err != nil {
		return nil, err
	}

	return &analysis, nil
}

func (r *Repository) FormAnalysisRequest(id uint) (*ds.AnalysisRequest, error) {
	var analysis ds.AnalysisRequest

	err := r.db.Where("analysis_request_id = ? AND analysis_request_status = 'черновик'", id).First(&analysis).Error
	if err != nil {
		return nil, err
	}

	if analysis.TextToAnalyse == "" {
		return nil, fmt.Errorf("текст для анализа не может быть пустым")
	}

	var count int64
	err = r.db.Model(&ds.AnalysisGenre{}).Where("analysis_request_id = ?", id).Count(&count).Error
	if err != nil {
		return nil, err
	}

	if count == 0 {
		return nil, fmt.Errorf("нельзя сформировать заявку без жанров")
	}

	analysis.AnalysisRequestStatus = "сформирован"
	analysis.FormedAt = sql.NullTime{Time: time.Now(), Valid: true}

	err = r.db.Save(&analysis).Error
	if err != nil {
		return nil, err
	}

	return &analysis, nil
}

func (r *Repository) DeleteAnalysisRequest(analysisID uint) error {
	return r.db.Exec("UPDATE analysis_requests SET analysis_request_status = 'удалён' WHERE analysis_request_id = ?", analysisID).Error
}

func (r *Repository) ProcessAnalysisRequest(id uint, moderatorID int, action string) (*ds.AnalysisRequestDTO, error) {
	var analysis ds.AnalysisRequest
	// 1. ПОИСК ЗАЯВКИ
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

	// 2. ПРОВЕРКА СТАТУСА
	if analysis.AnalysisRequestStatus != "сформирован" {
		return nil, fmt.Errorf("заявка не может быть обработана, так как ее текущий статус: %s. Требуется статус 'сформирован'", analysis.AnalysisRequestStatus)
	}

	// 3. ОБНОВЛЕНИЕ ModeratorID
	analysis.ModeratorID = sql.NullInt64{Int64: int64(moderatorID), Valid: true}

	switch action {
	case "complete":
		// ❌ ИСХОДНАЯ ЛОГИКА ВЫЧИСЛЕНИЙ УДАЛЕНА И ПЕРЕНЕСЕНА В Django
		// analysis.AnalysisRequestStatus остается "сформирован" (ожидание асинхронного сервиса)
		// CompletedAt не заполняется
		return nil, fmt.Errorf("действие 'complete' должно быть вызвано через ProcessAnalysisRequest с параметром 'start_analysis'")

	case "reject":
		analysis.AnalysisRequestStatus = "отклонён"
		analysis.CompletedAt = sql.NullTime{Time: time.Now(), Valid: true} // Завершается, если отклонена

	default:
		return nil, fmt.Errorf("недопустимое действие: %s. Допустимо 'complete' или 'reject'", action)
	}

	// 4. СОХРАНЕНИЕ В БАЗУ ДАННЫХ
	err = r.db.Save(&analysis).Error // Сохраняем обновленный ModeratorID и, возможно, статус
	if err != nil {
		return nil, err
	}

	var moderatorLogin string
	if analysis.ModeratorID.Valid {
		err = r.db.Table("users").
			Where("user_id = ?", analysis.ModeratorID.Int64).
			Select("login").
			Scan(&moderatorLogin).Error

		if err != nil {
			moderatorLogin = ""
		}
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

	if moderatorLogin != "" {
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

	return &dto, nil
}

// internal/app/repository/analysisrequestr.go

// ChangeStatusToProcessing меняет статус на 'на анализе' перед вызовом Django
func (r *Repository) ChangeStatusToProcessing(id uint, moderatorID int) (*ds.AnalysisRequestDTO, error) {
	var analysis ds.AnalysisRequest

	// 🔥 Проверяем, что заявка 'сформирована', и ищем ее
	err := r.db.Where("analysis_request_id = ? AND analysis_request_status = 'сформирован'", id).First(&analysis).Error

	if err != nil {
		return nil, fmt.Errorf("заявка не найдена или не находится в статусе 'сформирован': %w", err)
	}

	// Если заявка найдена и ее статус 'сформирован', мы НЕ МЕНЯЕМ ЕГО ЗДЕСЬ!
	// Мы просто присваиваем ID модератора.

	analysis.ModeratorID = sql.NullInt64{Int64: int64(moderatorID), Valid: true}

	// 🔥 УДАЛИТЕ эту строку: analysis.AnalysisRequestStatus = "на анализе"

	if err := r.db.Save(&analysis).Error; err != nil {
		return nil, err
	}

	// Возвращаем DTO
	return r.GetAnalysisRequestByID(int(id))
}

// UpdateAnalysisResults обрабатывает callback от Django
func (r *Repository) UpdateAnalysisResults(analysisID int, genreUpdates []ds.GenreUpdateData) error {
	tx := r.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 1. Обновляем AnalysisGenre
	for _, update := range genreUpdates {
		err := tx.Model(&ds.AnalysisGenre{}).
			Where("analysis_request_id = ? AND genre_id = ?", analysisID, update.GenreID).
			Updates(map[string]interface{}{
				"probability_percent": update.ProbabilityPercent,
			}).Error

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("ошибка обновления AnalysisGenre для жанра %d: %w", update.GenreID, err)
		}
	}

	// 2. Обновляем AnalysisRequest (статус на 'завершён' и время завершения)
	err := tx.Model(&ds.AnalysisRequest{}).
		Where("analysis_request_id = ?", analysisID).
		Updates(map[string]interface{}{
			"analysis_request_status": "завершён", // <-- Используем допустимый статус!
			"completed_at":            sql.NullTime{Time: time.Now(), Valid: true},
		}).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ошибка обновления статуса заявки: %w", err)
	}

	return tx.Commit().Error
}

// HandleAnalysisFailure откатывает статус на 'сформирован' в случае ошибки вызова Django
func (r *Repository) HandleAnalysisFailure(id uint, errorMessage string) error {
	logrus.Errorf("Обработка ошибки для заявки %d: %s", id, errorMessage)

	return r.db.Model(&ds.AnalysisRequest{}).
		Where("analysis_request_id = ?", id).
		Updates(map[string]interface{}{
			"analysis_request_status": "сформирован",
			// Если вы добавите поле error_message, можете обновить его здесь.
		}).Error
}
