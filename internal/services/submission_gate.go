package services

import (
	"algoforces/internal/domain"
	"time"
)

// CheckSubmissionGate is the pure, DB-free contest submission gate.
//
// It decides, from already-loaded rows, whether userID may submit to
// contest at time now. It returns one of the sentinel errors defined in
// domain/submission.go, or nil when the submission may proceed. The
// checks run in this order, so the first failing condition wins:
//
//  1. contest == nil -> domain.ErrContestNotFound
//  2. !contest.Visible -> domain.ErrContestNotVisible
//  3. not live (now before StartTime, after EndTime, or !IsActive)
//     -> domain.ErrContestNotLive. The window is inclusive:
//     StartTime <= now <= EndTime passes when IsActive.
//  4. no registration with Status == "registered", unless the user is
//     privileged for this contest -> domain.ErrNotRegistered
//  5. problem is not in the contest -> domain.ErrProblemNotInContest
//
// The registration argument must be the registration row for this
// user+contest (or nil when there is none); the caller is responsible
// for loading it with that scope. A registration whose Status is
// anything other than "registered" (e.g. "unregistered" after the user
// unregistered) does not satisfy the gate.
//
// Privileged bypass: a user is privileged only when they are the
// contest's creator (contest.CreatedBy) or listed in
// contest.ProblemSetters. Privilege bypasses ONLY the registration
// check (step 4); visibility, liveness and problem membership are
// still enforced for privileged users. This is deliberate: creators /
// problem setters should be able to test-submit without registering,
// but not into an invisible contest, outside the live window, or for
// a problem that is not in the contest.
//
// NOTE — intentional duplication (dedup follow-up): the
// visibility/liveness computation here (Visible, StartTime, EndTime,
// IsActive) duplicates the contest status logic owned by PR1 (contest
// lifecycle / start / end / status). PR2 must stay standalone and must
// not import any status helper from PR1, so the logic is duplicated
// here on purpose. After PR1 and PR2 are both merged, replace the
// inline computation below with the shared PR1 helper and keep this
// gate's ordering and sentinel errors.
func CheckSubmissionGate(contest *domain.Contest, registration *domain.ContestRegistration, problemInContest bool, userID string, now time.Time) error {
	if contest == nil {
		return domain.ErrContestNotFound
	}
	if !contest.Visible {
		return domain.ErrContestNotVisible
	}
	if !contest.IsActive || now.Before(contest.StartTime) || now.After(contest.EndTime) {
		return domain.ErrContestNotLive
	}
	if !isContestPrivileged(contest, userID) {
		if registration == nil || registration.Status != "registered" {
			return domain.ErrNotRegistered
		}
	}
	if !problemInContest {
		return domain.ErrProblemNotInContest
	}
	return nil
}

// isContestPrivileged reports whether userID is the contest creator or
// one of the contest's problem setters. No repository lookup is needed:
// both facts live on the Contest row.
func isContestPrivileged(contest *domain.Contest, userID string) bool {
	if contest.CreatedBy == userID {
		return true
	}
	for _, setterID := range contest.ProblemSetters {
		if setterID == userID {
			return true
		}
	}
	return false
}
