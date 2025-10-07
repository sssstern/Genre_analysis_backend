package repository

import (
	"lab3/internal/app/ds"
)

func (r *Repository) UpdateAnalysisGenre(userID, genreID int, comment string, probability int) error {
	analysis, err := r.GetCurrentAnalysis(userID)
	if err != nil {
		return err
	}

	var ag ds.AnalysisGenre
	err = r.db.Where("analysis_request_id = ? AND genre_id = ?", analysis.AnalysisRequestID, genreID).First(&ag).Error
	if err != nil {
		return err
	}

	ag.CommentToRequest = comment
	ag.ProbabilityPercent = probability

	return r.db.Save(&ag).Error
}

func (r *Repository) RemoveGenreFromAnalysis(userID, genreID int) error {
	analysis, err := r.GetCurrentAnalysis(userID)
	if err != nil {
		return err
	}

	return r.db.Where("analysis_request_id = ? AND genre_id = ?", analysis.AnalysisRequestID, genreID).Delete(&ds.AnalysisGenre{}).Error
}
