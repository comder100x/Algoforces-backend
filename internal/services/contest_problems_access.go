package services

import (
	"algoforces/internal/domain"
	"time"
)

// CanViewContestProblemsForParticipant is the pure access decision for the
// participant-facing contest-problems view. It performs no I/O so it can be
// unit-tested with plain structs and no database.
//
// Access rules, evaluated in order:
//
//  1. Admins and problem-setters (by role) may view anytime.
//  2. The contest creator and the problem-setters listed on the contest may
//     view anytime.
//  3. A registered participant (registration != nil and Status ==
//     "registered") may view only during the live window, computed directly
//     from the Contest fields: StartTime <= now <= EndTime and IsActive.
//  4. Everyone else may not view.
//
// The caller must pass the registration row for this user and contest (or
// nil when there is none); the helper does not re-check that the
// registration belongs to the user or contest.
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
