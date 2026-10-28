package services

import (
	"algoforces/internal/domain"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// StandingsService computes contest standings from submission data.
type StandingsService struct {
	standingsRepo domain.StandingsRepository
	userRepo      domain.UserRepository
	contestRepo   domain.ContestRepository
}

// NewStandingsService creates a StandingsService.
func NewStandingsService(standingsRepo domain.StandingsRepository, userRepo domain.UserRepository, contestRepo domain.ContestRepository) domain.StandingsUseCase {
	return &StandingsService{
		standingsRepo: standingsRepo,
		userRepo:      userRepo,
		contestRepo:   contestRepo,
	}
}

// GetStandings returns the leaderboard. Before the contest starts, only
// admins, problem setters and listed setters may view it.
func (s *StandingsService) GetStandings(ctx context.Context, contestID string, userID string, role string) (*domain.StandingsResponse, error) {
	contest, err := s.contestRepo.GetByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrContestNotFound
		}
		return nil, err
	}
	if contest == nil {
		return nil, domain.ErrContestNotFound
	}

	status := contestStatusAt(contest, time.Now())
	if status == domain.ContestStatusUpcoming && !canViewUpcomingStandings(contest, userID, role) {
		log.Warn().Str("contest_id", contestID).Str("user_id", userID).Str("role", role).Msg("Standings requested before contest start by participant")
		return nil, domain.ErrStandingsForbidden
	}

	accepted, err := s.standingsRepo.GetAcceptedSubmissionsForContest(ctx, contestID)
	if err != nil {
		return nil, err
	}

	counts, err := s.standingsRepo.GetSubmissionCountsForContest(ctx, contestID)
	if err != nil {
		return nil, err
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

// usernamesForStandings resolves the username of every user who has
// at least one submission in the contest, from the accepted
// submissions and the attempt counts together.
func (s *StandingsService) usernamesForStandings(ctx context.Context, submissions []domain.Submission, counts map[domain.SubmissionCountKey]int) (map[string]string, error) {
	userIDs := make(map[string]struct{})
	for _, submission := range submissions {
		userIDs[submission.UserId] = struct{}{}
	}
	for key := range counts {
		userIDs[key.UserID] = struct{}{}
	}

	ordered := make([]string, 0, len(userIDs))
	for userID := range userIDs {
		ordered = append(ordered, userID)
	}
	sort.Strings(ordered)

	usernames := make(map[string]string, len(ordered))
	for _, userID := range ordered {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user %s: %w", userID, err)
		}
		usernames[userID] = user.Username
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
		sort.Slice(problemScores, func(i, j int) bool {
			return problemScores[i].ProblemID < problemScores[j].ProblemID
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

	sort.Slice(ranked, func(i, j int) bool {
		a, b := ranked[i], ranked[j]
		if a.entry.TotalScore != b.entry.TotalScore {
			return a.entry.TotalScore > b.entry.TotalScore
		}
		if (a.entry.LastAcceptedAt == nil) != (b.entry.LastAcceptedAt == nil) {
			return a.entry.LastAcceptedAt != nil
		}
		if a.entry.LastAcceptedAt != nil && b.entry.LastAcceptedAt != nil && !a.entry.LastAcceptedAt.Equal(*b.entry.LastAcceptedAt) {
			return a.entry.LastAcceptedAt.Before(*b.entry.LastAcceptedAt)
		}
		if a.totalAttempts != b.totalAttempts {
			return a.totalAttempts < b.totalAttempts
		}
		if a.entry.Username != b.entry.Username {
			return a.entry.Username < b.entry.Username
		}
		return a.entry.UserID < b.entry.UserID
	})

	entries := make([]domain.StandingsEntry, 0, len(ranked))
	for _, item := range ranked {
		entries = append(entries, item.entry)
	}
	return entries
}
