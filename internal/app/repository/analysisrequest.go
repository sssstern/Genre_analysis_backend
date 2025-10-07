package repository

import (
	"database/sql"
	"fmt"
	"lab3/internal/app/ds"
	"lab3/internal/app/service"
	"time"
)

func (r *Repository) GetCurrentAnalysisInfo(userID int) (currentAnalysisID int, count int64, err error) {
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
	var analysis ds.AnalysisRequest
	err := r.db.Where("creator_id = ? AND analysis_request_status = 'черновик'", userID).
		Preload("Genres").
		Preload("Genres.Genre").
		First(&analysis).Error

	if err != nil {
		if err.Error() == "record not found" {
			newAnalysis := &ds.AnalysisRequest{
				AnalysisRequestStatus: "черновик",
				TextToAnalyse:         "Старинный лесной лагерь казался тому возрасту каким-то райским местом. Но трудный подъем на вершину холма давался с огромным усилием. Привал был наградой. Мы с братцем сидели на крыльце походной палаты, поглядывая на котелок с едой. Он еле мог передвигать локоть от боли. Я готов был бороться с любым, кто посмеет нас рассудить или назвать убогими. Вдруг из-за забора выскочил рыжий кот, его лапа безнадежно торчала из ветки старого дуба. Он дико вопил. Мы в один миг вскочили. Братец, забыв про боль, махнул палкой: «Надо его спуститься!» Это был единственный способ. Я полез вверх. Высота была приличной. Девчонка из соседнего отряда, румяная девка, совала мне в карман халата пакет с молоком, шепча ласково: «Осторожнее!» Я кивнул, стараясь не смотреть вниз. Кот, почуяв свободу, выпустил когти. Я протянул руку, он царапнул мне ладонь до крови, но я сумел его схватить. Снизу послышался шепот и одобрительные выкрикивать. Спускаясь, я чувствовал, как по спине течет пот. На земле братец обнял меня, а рыжий зверь тут же прыгать к миске с молоком. Старушка-повариха усмехнулась: «Не зря старались». Мы перебили её, сказав одновременно: «Нельзя было допустить, чтобы животное страдало». Вечером в гостиной у костра та девчонка подошла, улыбнулась и поцеловала меня в щеку. Я вспыхнул, а братец засмеялся. В темноте фонарь светился мягким светом, а в воздухе витало что-то новое, большее, чем просто дружба. Мы помолчали, глядя на звезды. Казалось, сама ночь открылась нам навстречу.",
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

func (r *Repository) GetAnalysisRequests(status string, startDate, endDate time.Time) ([]ds.AnalysisRequestDTO, error) {
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

func (r *Repository) FormAnalysisRequest(id uint) error {
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

func (r *Repository) DeleteAnalysisRequest(analysisID uint) error {
	return r.db.Exec("UPDATE analysis_requests SET analysis_request_status = 'удалён' WHERE analysis_request_id = ?", analysisID).Error
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

	switch action {
	case "complete":
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

	case "reject":
		analysis.AnalysisRequestStatus = "отклонён"

	default:
		return nil, fmt.Errorf("недопустимое действие: %s. Допустимо 'complete' или 'reject'", action)
	}

	err = r.db.Save(&analysis).Error
	if err != nil {
		return nil, err
	}

	var moderatorLogin string
	if analysis.ModeratorID.Valid {
		// Прямой запрос логина пользователя по ID
		err = r.db.Table("users").
			Where("user_id = ?", analysis.ModeratorID.Int64).
			Select("login").
			Scan(&moderatorLogin).Error

		if err != nil {
			// Если модератор не найден, это не должно ломать логику
			moderatorLogin = ""
		}
	}

	// Формирование DTO
	dto := ds.AnalysisRequestDTO{
		AnalysisRequestID:     analysis.AnalysisRequestID,
		AnalysisRequestStatus: analysis.AnalysisRequestStatus,
		CreatedAt:             analysis.CreatedAt,
		CreatorLogin:          analysis.Creator.Login, // Creator актуален, т.к. не менялся
		TextToAnalyse:         analysis.TextToAnalyse,
	}

	if analysis.FormedAt.Valid {
		dto.FormedAt = &analysis.FormedAt.Time
	}
	if analysis.CompletedAt.Valid {
		dto.CompletedAt = &analysis.CompletedAt.Time
	}

	// 💡 Присваиваем актуальный логин
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
