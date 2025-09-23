package ds

import (
	"database/sql"
	"time"
)

type AnalysisRequest struct {
	AnalysisRequestID     int           `gorm:"primaryKey;column:analysis_request_id"`
	AnalysisRequestStatus string        `gorm:"type:varchar(20);not null;default:'черновик';column:analysis_request_status"`
	CreatedAt             time.Time     `gorm:"not null;column:created_at"`
	CreatorID             int           `gorm:"not null;column:creator_id"`
	FormedAt              sql.NullTime  `gorm:"column:formed_at"`
	CompletedAt           sql.NullTime  `gorm:"column:completed_at"`
	ModeratorID           sql.NullInt64 `gorm:"column:moderator_id"`
	TextToAnalyse         string        `gorm:"type:text;not null;column:text_to_analyse"`

	Creator   User            `gorm:"foreignKey:CreatorID;references:UserID"`
	Moderator User            `gorm:"foreignKey:ModeratorID;references:UserID"`
	Genres    []AnalysisGenre `gorm:"foreignKey:AnalysisRequestID;references:AnalysisRequestID"`
}
