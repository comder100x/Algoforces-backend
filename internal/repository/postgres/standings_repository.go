package postgres

import (
	"algoforces/internal/domain"
	"context"

	"gorm.io/gorm"
)

// NewStandingsRepository returns a StandingsRepository backed by the same
// submissionRepository struct; no new table or migration is involved.
func NewStandingsRepository(db *gorm.DB) domain.StandingsRepository {
	return &submissionRepository{
		db: db,
	}
}

// GetAcceptedSubmissionsForContest returns every submission in the
// contest whose verdict is Accepted, oldest first.
func (r *submissionRepository) GetAcceptedSubmissionsForContest(ctx context.Context, contestID string) ([]domain.Submission, error) {
	var submissions []domain.Submission
	err := r.db.WithContext(ctx).
		Where("contest_id = ? AND verdict = ?", contestID, string(domain.VerdictAccepted)).
		Order("submitted_at ASC").
		Find(&submissions).Error
	if err != nil {
		return nil, err
	}
	return submissions, nil
}

// GetSubmissionCountsForContest returns, for every user/problem pair
// with at least one submission in the contest, the number of
// submissions that user made on that problem (all verdicts).
func (r *submissionRepository) GetSubmissionCountsForContest(ctx context.Context, contestID string) (map[domain.SubmissionCountKey]int, error) {
	type submissionCountRow struct {
		UserID    string `gorm:"column:user_id"`
		ProblemID string `gorm:"column:problem_id"`
		Count     int    `gorm:"column:count"`
	}

	var rows []submissionCountRow
	err := r.db.WithContext(ctx).
		Model(&domain.Submission{}).
		Select("user_id, problem_id, COUNT(*) as count").
		Where("contest_id = ?", contestID).
		Group("user_id, problem_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[domain.SubmissionCountKey]int, len(rows))
	for _, row := range rows {
		counts[domain.SubmissionCountKey{UserID: row.UserID, ProblemID: row.ProblemID}] = row.Count
	}
	return counts, nil
}
