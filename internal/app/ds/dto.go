package ds

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
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
	GenresCompletedCount  int                `json:"genres_completed_count"`
}

type AnalysisGenreDTO struct {
	GenreID            int    `json:"GenreID"`
	GenreName          string `json:"GenreName"`
	GenreImageURL      string `json:"GenreImageURL"`
	GenreKeywords      string
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
	UserID int      `json:"UserID"`
	Login  string   `json:"Login"`
	Role   UserRole `json:"Role"` // Теперь строковое поле
}

type ChangeUserDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type UserRole string

const (
	RoleGuest     UserRole = "guest"
	RoleCreator   UserRole = "creator" // Обычный пользователь
	RoleModerator UserRole = "moderator"
)

// JWTClaims определяет данные, которые мы храним в токене
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID int      `json:"user_id"`
	Role   UserRole `json:"role"`
}

type AuthResponseDTO struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"` // Unix timestamp истечения
}

type GenreUpdateData struct {
	GenreID            int `json:"genre_id"`
	ProbabilityPercent int `json:"probability_percent"`
}

// AnalysisUpdateFromDjango - DTO для приема данных из Django
type AnalysisUpdateFromDjango struct {
	AnalysisRequestID int               `json:"analysis_request_id"`
	SecretKey         string            `json:"secret_key"`
	AnalysisGenreData []GenreUpdateData `json:"analysis_genre_data"`
}
