package ds

type Genre struct {
	GenreID       int    `gorm:"primaryKey;column:genre_id"`
	GenreName     string `gorm:"type:varchar(100);not null;column:genre_name"`
	GenreImageURL string `gorm:"type:varchar(255);column:genre_image_url"`
	GenreKeywords string `gorm:"type:text;column:genre_keywords"`
	IsDeleted     bool   `gorm:"type:boolean;default:false;column:is_deleted"`
}
