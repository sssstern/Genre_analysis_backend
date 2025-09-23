package repository

import (
	"fmt"
	"lab2/internal/app/ds"
)

func (r *Repository) GetGenres() ([]ds.Genre, error) {
	var genres []ds.Genre
	err := r.db.Where("is_deleted = false").Find(&genres).Error
	// обязательно проверяем ошибки, и если они появились - передаем выше, то есть хендлеру
	if err != nil {
		return nil, err
	}
	if len(genres) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	return genres, nil
}

func (r *Repository) GetGenre(id int) (ds.Genre, error) {
	genre := ds.Genre{}
	err := r.db.Where("genre_id = ? AND is_deleted = false", id).First(&genre).Error
	if err != nil {
		return ds.Genre{}, err
	}
	return genre, nil
}

func (r *Repository) GetGenresByTitle(title string) ([]ds.Genre, error) {
	var genres []ds.Genre
	err := r.db.Where("genre_name ILIKE ? AND is_deleted = false", "%"+title+"%").Find(&genres).Error
	if err != nil {
		return nil, err
	}
	return genres, nil
}
