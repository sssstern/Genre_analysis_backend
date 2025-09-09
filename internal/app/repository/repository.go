package repository

import (
	"fmt"
	"strings"
	"time"
)

type Repository struct {
}

func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

type Genre struct {
	GenreID    int
	GenreName  string
	ImageURL   string
	GenrePrice float64
	Keywords   string
}

type GenreAnalysisRequest struct {
	GenreAnalysisRequestID int
	CreatedAt              time.Time
	Genres                 []AnalysisGenreItem
	Status                 string
}

type AnalysisGenreItem struct {
	GenreID int
	Comment string
}

func (r *Repository) GetGenres() ([]Genre, error) {
	genres := []Genre{
		{
			GenreID:    1,
			GenreName:  "Хроника",
			ImageURL:   "http://localhost:9000/genreanalysis/Khronika.png",
			GenrePrice: 1000.00,
			Keywords:   "жизнь, лицо, домой, времени, голос, стоял",
		},
		{
			GenreID:    2,
			GenreName:  "Житие",
			ImageURL:   "http://localhost:9000/genreanalysis/Zhitie.png",
			GenrePrice: 1129.00,
			Keywords:   "опять, совсем, спросил, стал, дело, почти, сказала,  день, лет, голову, сразу, руки",
		},
		{
			GenreID:    3,
			GenreName:  "Договор",
			ImageURL:   "http://localhost:9000/genreanalysis/Dogovor.png",
			GenrePrice: 962.00,
			Keywords:   "несколько, вместе, году, вообще, совершенно, людей, говорит, мог, кажется, два, сразу",
		},
		{
			GenreID:    4,
			GenreName:  "Поэма",
			GenrePrice: 1013.00,
			ImageURL:   "http://localhost:9000/genreanalysis/Poema.png",
			Keywords:   "бог, быть, сын, были, одиссей, а, долго, отец, дом, давно, мог, старик, лет, опять",
		},
	}

	if len(genres) == 0 {
		return nil, fmt.Errorf("массив жанров пустой")
	}

	return genres, nil
}

func (r *Repository) GetGenre(id int) (Genre, error) {
	// Получаем все жанры
	genres, err := r.GetGenres()
	if err != nil {
		return Genre{}, err
	}

	// Ищем жанр по ID
	for _, genre := range genres {
		if genre.GenreID == id {
			return genre, nil
		}
	}

	return Genre{}, fmt.Errorf("жанр с ID %d не найден", id)
}

func (r *Repository) GetGenresByTitle(query string) ([]Genre, error) {
	// Получаем все жанры
	genres, err := r.GetGenres()
	if err != nil {
		return []Genre{}, err
	}

	var result []Genre

	// Ищем по названию жанра
	for _, genre := range genres {
		if strings.Contains(strings.ToLower(genre.GenreName), strings.ToLower(query)) {
			result = append(result, genre)
		}
	}

	return result, nil
}

func (r *Repository) GetCurrentAnalysis() *GenreAnalysisRequest {
	currentAnalysis := &GenreAnalysisRequest{
		GenreAnalysisRequestID: 1, //Захардкорженная заявка всвязи
		CreatedAt:              time.Now(),
		Status:                 "формируется",
		Genres: []AnalysisGenreItem{
			{GenreID: 1, Comment: "Срочный анализ"},
			{GenreID: 3, Comment: "Проверить договор"},
			{GenreID: 2},
		},
	}
	return currentAnalysis
}

func (r *Repository) GetAnalysisCount() int {
	currentAnalysis := r.GetCurrentAnalysis()
	return len(currentAnalysis.Genres)
}
