package ds

type AnalysisGenre struct {
	AnalysisRequestID  int    `gorm:"primaryKey;not null;uniqueIndex:idx_analysis_genre;column:analysis_request_id"`
	GenreID            int    `gorm:"primaryKey;not null;uniqueIndex:idx_analysis_genre;column:genre_id"`
	CommentToRequest   string `gorm:"type:text;column:comment_to_request"`
	ProbabilityPercent int    `gorm:"not null;default:0;column:probability_percent"`

	AnalysisRequest AnalysisRequest `gorm:"foreignKey:AnalysisRequestID;references:AnalysisRequestID"`
	Genre           Genre           `gorm:"foreignKey:GenreID;references:GenreID"`
}
