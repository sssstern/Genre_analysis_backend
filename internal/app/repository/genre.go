// repository/genre.go
package repository

import (
	"fmt"
	"lab3/internal/app/ds"
)

func (r *Repository) GetGenres() ([]ds.GenreDTO, error) {
	var genres []ds.Genre
	err := r.db.Where("is_deleted = false").Find(&genres).Error
	if err != nil {
		return nil, err
	}
	if len(genres) == 0 {
		return nil, fmt.Errorf("массив пустой")
	}

	genreDTOs := make([]ds.GenreDTO, len(genres))
	for i, g := range genres {
		genreDTOs[i] = ds.GenreDTO{
			GenreID:       g.GenreID,
			GenreName:     g.GenreName,
			GenreImageURL: g.GenreImageURL,
			GenreKeywords: g.GenreKeywords,
		}
	}
	return genreDTOs, nil
}

func (r *Repository) GetGenresByTitle(title string) ([]ds.GenreDTO, error) {
	var genres []ds.Genre
	err := r.db.Where("genre_name ILIKE ? AND is_deleted = false", "%"+title+"%").Find(&genres).Error
	if err != nil {
		return nil, err
	}
	genreDTOs := make([]ds.GenreDTO, len(genres))
	for i, g := range genres {
		genreDTOs[i] = ds.GenreDTO{
			GenreID:       g.GenreID,
			GenreName:     g.GenreName,
			GenreImageURL: g.GenreImageURL,
			GenreKeywords: g.GenreKeywords,
		}
	}
	return genreDTOs, nil
}

func (r *Repository) GetGenre(id int) (ds.GenreDTO, error) {
	var genre ds.Genre
	err := r.db.Where("genre_id = ? AND is_deleted = false", id).First(&genre).Error
	if err != nil {
		return ds.GenreDTO{}, err
	}
	return ds.GenreDTO{
		GenreID:       genre.GenreID,
		GenreName:     genre.GenreName,
		GenreImageURL: genre.GenreImageURL,
		GenreKeywords: genre.GenreKeywords,
	}, nil
}

func (r *Repository) CreateGenre(genreDTO ds.GenreDTO) (*ds.GenreDTO, error) {
	genre := ds.Genre{
		GenreName:     genreDTO.GenreName,
		GenreImageURL: genreDTO.GenreImageURL,
		GenreKeywords: genreDTO.GenreKeywords,
		IsDeleted:     false,
	}
	err := r.db.Create(&genre).Error
	if err != nil {
		return nil, err
	}
	return &ds.GenreDTO{
		GenreID:       genre.GenreID,
		GenreName:     genre.GenreName,
		GenreImageURL: genre.GenreImageURL,
		GenreKeywords: genre.GenreKeywords,
	}, nil
}

func (r *Repository) UpdateGenre(id int, genreUpdates ds.UpdateGenreRequestDTO) (*ds.GenreDTO, error) {
	var genre ds.Genre
	err := r.db.Where("genre_id = ? AND is_deleted = false", id).First(&genre).Error
	if err != nil {
		return nil, err
	}

	if genreUpdates.GenreName != "" {
		genre.GenreName = genreUpdates.GenreName
	}
	if genreUpdates.GenreKeywords != "" {
		genre.GenreKeywords = genreUpdates.GenreKeywords
	}

	err = r.db.Save(&genre).Error
	if err != nil {
		return nil, err
	}

	return &ds.GenreDTO{
		GenreID:       genre.GenreID,
		GenreName:     genre.GenreName,
		GenreImageURL: genre.GenreImageURL,
		GenreKeywords: genre.GenreKeywords,
	}, nil
}

func (r *Repository) DeleteGenre(id int) error {
	return r.db.Model(&ds.Genre{}).Where("genre_id = ?", id).Update("is_deleted", true).Error
}

func (r *Repository) UpdateGenreImage(id int, imageURL string) (*ds.GenreDTO, error) {
	var genre ds.Genre
	err := r.db.Where("genre_id = ? AND is_deleted = false", id).First(&genre).Error
	if err != nil {
		return nil, err
	}

	genre.GenreImageURL = imageURL

	err = r.db.Save(&genre).Error
	if err != nil {
		return nil, err
	}

	return &ds.GenreDTO{
		GenreID:       genre.GenreID,
		GenreName:     genre.GenreName,
		GenreImageURL: genre.GenreImageURL,
		GenreKeywords: genre.GenreKeywords,
	}, nil
}
