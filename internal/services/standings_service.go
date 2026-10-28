package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
)

type standingsService struct {
	standingsRepo domain.StandingsRepository
	usernameRepo  domain.UsernameRepository
	contestRepo   domain.ContestRepository
}

// NewStandingsService creates a standings use case.
func NewStandingsService(standingsRepo domain.StandingsRepository, usernameRepo domain.UsernameRepository, contestRepo domain.ContestRepository) domain.StandingsUseCase {
	return &standingsService{
		standingsRepo: standingsRepo,
		usernameRepo:  usernameRepo,
		contestRepo:   contestRepo,
	}
}

// GetStandings returns the leaderboard. Before the contest starts, only
// admins, problem setters and listed setters may view it.
func (s *standingsService) GetStandings(ctx context.Context, contestID string, userID string, role string) (*domain.StandingsResponse, error) {
	if s == nil || s.standingsRepo == nil || s.usernameRepo == nil || s.contestRepo == nil {
		return nil, errors.New("standings service is not initialised")
	}

	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrContestNotFound
		}
		return nil, fmt.Errorf("get contest %s: %w", contestID, err)
	}
	if contest == nil {
		return nil, domain.ErrContestNotFound
	}

	status := contestStatusAt(contest, time.Now())
	if status == domain.ContestStatusUpcoming && !canViewUpcomingStandings(contest, userID, role) {
		return nil, domain.ErrStandingsForbidden
	}

	accepted, err := s.standingsRepo.GetAcceptedSubmissionsForContest(ctx, contestID)
	if err != nil {
		return nil, fmt.Errorf("get accepted submissions for contest %s: %w", contestID, err)
	}

	counts, err := s.standingsRepo.GetSubmissionCountsForContest(ctx, contestID)
	if err != nil {
		return nil, fmt.Errorf("get submission counts for contest %s: %w", contestID, err)
	}

	usernames, err := s.usernamesForStandings(ctx, accepted, counts)
	if err != nil {
		return nil, err
	}

	entries := BuildStandingsEntries(accepted, usernames, counts)

	return &domain.StandingsResponse{
		ContestID:   contestID,
		Status:      status,
		Entries:     entries,
		GeneratedAt: time.Now(),
	}, nil
}

// contestStatusAt derives status from StartTime, EndTime and IsActive.
func contestStatusAt(contest *domain.Contest, now time.Time) string {
	if now.After(contest.EndTime) {
		return domain.ContestStatusEnded
	}
	if !contest.IsActive || now.Before(contest.StartTime) {
		return domain.ContestStatusUpcoming
	}
	return domain.ContestStatusLive
}

// canViewUpcomingStandings reports whether a caller may view the
// standings before the contest starts: admins, problem setters
// (by role), and users listed as problem setters on this contest.
func canViewUpcomingStandings(contest *domain.Contest, userID string, role string) bool {
	if role == "admin" || role == "problem-setter" {
		return true
	}
	for _, setterID := range contest.ProblemSetters {
		if setterID == userID {
			return true
		}
	}
	return false
}

// usernamesForStandings resolves usernames in one batch query. Users
// missing from the result fall back to their user ID.
func (s *standingsService) usernamesForStandings(ctx context.Context, submissions []domain.Submission, counts map[domain.SubmissionCountKey]int) (map[string]string, error) {
	userIDs := make(map[string]struct{})
	for _, submission := range submissions {
		userIDs[submission.UserId] = struct{}{}
	}
	for key := range counts {
		userIDs[key.UserID] = struct{}{}
	}
	if len(userIDs) == 0 {
		return map[string]string{}, nil
	}

	ordered := make([]string, 0, len(userIDs))
	for userID := range userIDs {
		ordered = append(ordered, userID)
	}
	slices.Sort(ordered)

	loaded, err := s.usernameRepo.GetUsernamesByIDs(ctx, ordered)
	if err != nil {
		return nil, fmt.Errorf("get usernames for standings: %w", err)
	}
	if loaded == nil {
		loaded = map[string]string{}
	}

	usernames := make(map[string]string, len(ordered))
	for _, userID := range ordered {
		username, ok := loaded[userID]
		if !ok || username == "" {
			username = userID
		}
		usernames[userID] = username
	}
	return usernames, nil
}

type standingsProblemAggregate struct {
	bestScore       int
	firstAcceptedAt *time.Time
	attempts        int
	hasAttempts     bool
	acceptedCount   int
}

type standingsUserAggregate struct {
	problems map[string]*standingsProblemAggregate
}

type rankedStandingsEntry struct {
	entry         domain.StandingsEntry
	totalAttempts int
}

// BuildStandingsEntries ranks users from submissions and attempt counts.
// It is pure; entries sort by total score desc, last-accepted asc, attempts asc, then username.
func BuildStandingsEntries(submissions []domain.Submission, usernames map[string]string, counts map[domain.SubmissionCountKey]int) []domain.StandingsEntry {
	users := make(map[string]*standingsUserAggregate)

	problemAggregate := func(userID, problemID string) *standingsProblemAggregate {
		user, ok := users[userID]
		if !ok {
			user = &standingsUserAggregate{problems: make(map[string]*standingsProblemAggregate)}
			users[userID] = user
		}
		problem, ok := user.problems[problemID]
		if !ok {
			problem = &standingsProblemAggregate{}
			user.problems[problemID] = problem
		}
		return problem
	}

	// Attempt counts first, so users with no Accepted submission
	// still get an entry and their failed attempts are counted.
	for key, count := range counts {
		if count <= 0 {
			continue
		}
		problem := problemAggregate(key.UserID, key.ProblemID)
		problem.attempts = count
		problem.hasAttempts = true
	}

	// Accepted submissions only: best score and first-accepted time.
	for _, submission := range submissions {
		if submission.Verdict != string(domain.VerdictAccepted) {
			continue
		}
		problem := problemAggregate(submission.UserId, submission.ProblemID)
		problem.acceptedCount++
		if submission.Score > problem.bestScore {
			problem.bestScore = submission.Score
		}
		if problem.firstAcceptedAt == nil || submission.SubmittedAt.Before(*problem.firstAcceptedAt) {
			submittedAt := submission.SubmittedAt
			problem.firstAcceptedAt = &submittedAt
		}
	}

	ranked := make([]rankedStandingsEntry, 0, len(users))
	for userID, user := range users {
		problemScores := make([]domain.StandingsProblemScore, 0, len(user.problems))
		totalScore := 0
		solvedCount := 0
		totalAttempts := 0
		var lastAcceptedAt *time.Time
		for problemID, problem := range user.problems {
			attempts := problem.attempts
			if !problem.hasAttempts {
				attempts = problem.acceptedCount
			}
			totalAttempts += attempts
			totalScore += problem.bestScore
			if problem.bestScore > 0 {
				solvedCount++
				if problem.firstAcceptedAt != nil && (lastAcceptedAt == nil || problem.firstAcceptedAt.After(*lastAcceptedAt)) {
					firstAcceptedAt := *problem.firstAcceptedAt
					lastAcceptedAt = &firstAcceptedAt
				}
			}
			problemScores = append(problemScores, domain.StandingsProblemScore{
				ProblemID:       problemID,
				BestScore:       problem.bestScore,
				FirstAcceptedAt: problem.firstAcceptedAt,
				Attempts:        attempts,
			})
		}
		slices.SortFunc(problemScores, func(a, b domain.StandingsProblemScore) int {
			return strings.Compare(a.ProblemID, b.ProblemID)
		})

		username := usernames[userID]
		if username == "" {
			username = userID
		}
		ranked = append(ranked, rankedStandingsEntry{
			entry: domain.StandingsEntry{
				UserID:         userID,
				Username:       username,
				TotalScore:     totalScore,
				SolvedCount:    solvedCount,
				LastAcceptedAt: lastAcceptedAt,
				ProblemScores:  problemScores,
			},
			totalAttempts: totalAttempts,
		})
	}

	slices.SortFunc(ranked, compareRankedStandingsEntries)

	entries := make([]domain.StandingsEntry, 0, len(ranked))
	for _, item := range ranked {
		entries = append(entries, item.entry)
	}
	return entries
}

func compareRankedStandingsEntries(a, b rankedStandingsEntry) int {
	scoreDiffers := a.entry.TotalScore != b.entry.TotalScore
	if scoreDiffers {
		if a.entry.TotalScore > b.entry.TotalScore {
			return -1
		}
		return 1
	}

	aHasLastAccepted := a.entry.LastAcceptedAt != nil
	bHasLastAccepted := b.entry.LastAcceptedAt != nil
	lastAcceptedPresenceDiffers := aHasLastAccepted != bHasLastAccepted
	if lastAcceptedPresenceDiffers {
		if aHasLastAccepted {
			return -1
		}
		return 1
	}

	lastAcceptedDiffers := aHasLastAccepted && bHasLastAccepted && !a.entry.LastAcceptedAt.Equal(*b.entry.LastAcceptedAt)
	if lastAcceptedDiffers {
		if a.entry.LastAcceptedAt.Before(*b.entry.LastAcceptedAt) {
			return -1
		}
		return 1
	}

	attemptsDiffer := a.totalAttempts != b.totalAttempts
	if attemptsDiffer {
		if a.totalAttempts < b.totalAttempts {
			return -1
		}
		return 1
	}

	usernameDiffers := a.entry.Username != b.entry.Username
	if usernameDiffers {
		return strings.Compare(a.entry.Username, b.entry.Username)
	}
	return strings.Compare(a.entry.UserID, b.entry.UserID)
}
