package ds

import (
	"time"
)

type AnalysisRequestDTO struct {
	AnalysisRequestID     int                `json:"AnalysisRequestID"`
	AnalysisRequestStatus string             `json:"AnalysisRequestStatus"`
	CreatedAt             time.Time          `json:"CreatedAt"`
	CreatorLogin          string             `json:"CreatorLogin"`
	FormedAt              *time.Time         `json:"FormedAt,omitempty"`
	CompletedAt           *time.Time         `json:"CompletedAt,omitempty"`
	ModeratorLogin        *string            `json:"ModeratorLogin,omitempty"`
	TextToAnalyse         string             `json:"TextToAnalyse"`
	Genres                []AnalysisGenreDTO `json:"Genres"`
}

type AnalysisGenreDTO struct {
	GenreID            int    `json:"GenreID"`
	GenreName          string `json:"GenreName"`
	GenreImageURL      string `json:"GenreImageURL"`
	CommentToRequest   string `json:"CommentToRequest"`
	ProbabilityPercent int    `json:"ProbabilityPercent"`
}

type UpdateAnalysisRequestDTO struct {
	TextToAnalyse string `json:"TextToAnalyse"`
}

type UpdateGenreRequestDTO struct {
	CommentToRequest   string `json:"comment_to_request"`
	ProbabilityPercent int    `json:"probability_percent"`
}

type GenreDTO struct {
	GenreID       int    `json:"GenreID"`
	GenreName     string `json:"GenreName"`
	GenreImageURL string `json:"GenreImageURL"`
	GenreKeywords string `json:"GenreKeywords"`
}

type UpdateGenreDTO struct {
	GenreName     string `json:"GenreName"`
	GenreKeywords string `json:"GenreKeywords"`
}

type UserDTO struct {
	UserID int    `json:"UserID"`
	Login  string `json:"Login"`
}

type ChangeUserDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
