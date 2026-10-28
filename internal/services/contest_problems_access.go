package services

import (
	"algoforces/internal/domain"
	"time"
)

// CanViewContestProblemsForParticipant is the pure access decision for the
// participant-facing view. Callers pass the registration row, or nil.
func CanViewContestProblemsForParticipant(contest *domain.Contest, userID, role string, registration *domain.ContestRegistration, now time.Time) bool {
	if contest == nil {
		return false
	}

	// Privileged viewers: anytime, regardless of the live window.
	if role == "admin" || role == "problem-setter" {
		return true
	}
	if userID != "" && userID == contest.CreatedBy {
		return true
	}
	for _, setterID := range contest.ProblemSetters {
		if setterID == userID {
			return true
		}
	}

	// Participants: only when registered, and only during the live window.
	if registration == nil || registration.Status != "registered" {
		return false
	}
	if !contest.IsActive {
		return false
	}
	if now.Before(contest.StartTime) || now.After(contest.EndTime) {
		return false
	}
	return true
}
