package services_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/services"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func standingsBaseTime() time.Time {
	return time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
}

func acceptedSubmission(userID, problemID string, score int, at time.Time) domain.Submission {
	return domain.Submission{
		UserId:      userID,
		ContestID:   "contest-1",
		ProblemID:   problemID,
		Verdict:     string(domain.VerdictAccepted),
		Score:       score,
		SubmittedAt: at,
	}
}

func rejectedSubmission(userID, problemID string, score int, at time.Time) domain.Submission {
	return domain.Submission{
		UserId:      userID,
		ContestID:   "contest-1",
		ProblemID:   problemID,
		Verdict:     string(domain.VerdictWrongAnswer),
		Score:       score,
		SubmittedAt: at,
	}
}

func verdictSubmission(userID, problemID string, verdict domain.VerdictStatus, score int, at time.Time) domain.Submission {
	return domain.Submission{
		UserId:      userID,
		ContestID:   "contest-1",
		ProblemID:   problemID,
		Verdict:     string(verdict),
		Score:       score,
		SubmittedAt: at,
	}
}

func submissionCountKey(userID, problemID string) domain.SubmissionCountKey {
	return domain.SubmissionCountKey{UserID: userID, ProblemID: problemID}
}

func standingsUserIDs(entries []domain.StandingsEntry) []string {
	ids := make([]string, 0, len(entries))
	for _, entry := range entries {
		ids = append(ids, entry.UserID)
	}
	return ids
}

func standingsEntryByUser(entries []domain.StandingsEntry, userID string) (domain.StandingsEntry, bool) {
	for _, entry := range entries {
		if entry.UserID == userID {
			return entry, true
		}
	}
	return domain.StandingsEntry{}, false
}

func standingsProblemScore(entry domain.StandingsEntry, problemID string) (domain.StandingsProblemScore, bool) {
	for _, problemScore := range entry.ProblemScores {
		if problemScore.ProblemID == problemID {
			return problemScore, true
		}
	}
	return domain.StandingsProblemScore{}, false
}

func TestBuildStandingsEntriesEmptyContest(t *testing.T) {
	entries := services.BuildStandingsEntries(nil, nil, nil)
	require.NotNil(t, entries)
	assert.Empty(t, entries)

	entries = services.BuildStandingsEntries([]domain.Submission{}, map[string]string{}, map[domain.SubmissionCountKey]int{})
	require.NotNil(t, entries)
	assert.Empty(t, entries)
}

func TestBuildStandingsEntriesOrder(t *testing.T) {
	base := standingsBaseTime()
	t1 := base
	t2 := base.Add(time.Hour)
	t3 := base.Add(2 * time.Hour)
	t4 := base.Add(3 * time.Hour)

	tests := []struct {
		name        string
		submissions []domain.Submission
		usernames   map[string]string
		counts      map[domain.SubmissionCountKey]int
		wantOrder   []string
		wantTotals  []int
		wantSolved  []int
		wantLast    []*time.Time
	}{
		{
			name: "total score descending",
			submissions: []domain.Submission{
				acceptedSubmission("alice", "p1", 100, t1),
				acceptedSubmission("bob", "p1", 50, t1),
			},
			usernames: map[string]string{"alice": "alice", "bob": "bob"},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 1,
				submissionCountKey("bob", "p1"):   1,
			},
			wantOrder:  []string{"alice", "bob"},
			wantTotals: []int{100, 50},
			wantSolved: []int{1, 1},
			wantLast:   []*time.Time{&t1, &t1},
		},
		{
			name: "best score is max accepted and first accepted is earliest",
			submissions: []domain.Submission{
				acceptedSubmission("alice", "p1", 50, t2),
				acceptedSubmission("alice", "p1", 100, t1),
				acceptedSubmission("alice", "p1", 80, t3),
			},
			usernames: map[string]string{"alice": "alice"},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 3,
			},
			wantOrder:  []string{"alice"},
			wantTotals: []int{100},
			wantSolved: []int{1},
			wantLast:   []*time.Time{&t1},
		},
		{
			name: "non-accepted submissions ignored for score but failed-only user still appears",
			submissions: []domain.Submission{
				rejectedSubmission("alice", "p1", 999, t1),
				acceptedSubmission("alice", "p1", 100, t2),
				rejectedSubmission("bob", "p1", 0, t1),
			},
			usernames: map[string]string{"alice": "alice", "bob": "bob"},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 2,
				submissionCountKey("bob", "p1"):   1,
			},
			wantOrder:  []string{"alice", "bob"},
			wantTotals: []int{100, 0},
			wantSolved: []int{1, 0},
			wantLast:   []*time.Time{&t2, nil},
		},
		{
			name: "tie broken by last accepted time",
			submissions: []domain.Submission{
				acceptedSubmission("alice", "p1", 100, t1),
				acceptedSubmission("bob", "p1", 100, t2),
			},
			usernames: map[string]string{"alice": "alice", "bob": "bob"},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 1,
				submissionCountKey("bob", "p1"):   1,
			},
			wantOrder:  []string{"alice", "bob"},
			wantTotals: []int{100, 100},
			wantSolved: []int{1, 1},
			wantLast:   []*time.Time{&t1, &t2},
		},
		{
			name: "tie broken by latest first-accepted among solved problems",
			submissions: []domain.Submission{
				acceptedSubmission("alice", "p1", 100, t1),
				acceptedSubmission("alice", "p2", 100, t4),
				acceptedSubmission("bob", "p1", 100, t2),
				acceptedSubmission("bob", "p2", 100, t3),
			},
			usernames: map[string]string{"alice": "alice", "bob": "bob"},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 1,
				submissionCountKey("alice", "p2"): 1,
				submissionCountKey("bob", "p1"):   1,
				submissionCountKey("bob", "p2"):   1,
			},
			wantOrder:  []string{"bob", "alice"},
			wantTotals: []int{200, 200},
			wantSolved: []int{2, 2},
			wantLast:   []*time.Time{&t3, &t4},
		},
		{
			name: "tie broken by total attempts ascending",
			submissions: []domain.Submission{
				acceptedSubmission("alice", "p1", 100, t1),
				acceptedSubmission("bob", "p1", 100, t1),
			},
			usernames: map[string]string{"alice": "alice", "bob": "bob"},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 3,
				submissionCountKey("bob", "p1"):   1,
			},
			wantOrder:  []string{"bob", "alice"},
			wantTotals: []int{100, 100},
			wantSolved: []int{1, 1},
			wantLast:   []*time.Time{&t1, &t1},
		},
		{
			name: "tie broken by username ascending",
			submissions: []domain.Submission{
				acceptedSubmission("user-a", "p1", 100, t1),
				acceptedSubmission("user-b", "p1", 100, t1),
			},
			usernames: map[string]string{"user-a": "bob", "user-b": "alice"},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("user-a", "p1"): 1,
				submissionCountKey("user-b", "p1"): 1,
			},
			wantOrder:  []string{"user-b", "user-a"},
			wantTotals: []int{100, 100},
			wantSolved: []int{1, 1},
			wantLast:   []*time.Time{&t1, &t1},
		},
		{
			name: "zero-score accepted submission is not solved",
			submissions: []domain.Submission{
				acceptedSubmission("alice", "p1", 0, t1),
			},
			usernames: map[string]string{"alice": "alice"},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 1,
			},
			wantOrder:  []string{"alice"},
			wantTotals: []int{0},
			wantSolved: []int{0},
			wantLast:   []*time.Time{nil},
		},
		{
			name:        "zero and negative counts create no entry",
			submissions: nil,
			usernames:   map[string]string{},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 0,
				submissionCountKey("bob", "p1"):   -2,
			},
			wantOrder:  []string{},
			wantTotals: []int{},
			wantSolved: []int{},
			wantLast:   []*time.Time{},
		},
		{
			name:        "counts-only user appears with zero score",
			submissions: nil,
			usernames:   map[string]string{"carol": "carol"},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("carol", "p1"): 2,
			},
			wantOrder:  []string{"carol"},
			wantTotals: []int{0},
			wantSolved: []int{0},
			wantLast:   []*time.Time{nil},
		},
		{
			name: "full tie is broken by user id",
			submissions: []domain.Submission{
				acceptedSubmission("user-b", "p1", 100, t1),
				acceptedSubmission("user-a", "p1", 100, t1),
			},
			usernames: map[string]string{"user-a": "same", "user-b": "same"},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("user-a", "p1"): 1,
				submissionCountKey("user-b", "p1"): 1,
			},
			wantOrder:  []string{"user-a", "user-b"},
			wantTotals: []int{100, 100},
			wantSolved: []int{1, 1},
			wantLast:   []*time.Time{&t1, &t1},
		},
		{
			name: "non-accepted verdicts contribute no score",
			submissions: []domain.Submission{
				verdictSubmission("alice", "p1", domain.VerdictTimeLimitExceeded, 999, t1),
				verdictSubmission("alice", "p1", domain.VerdictCompilationError, 999, t2),
				verdictSubmission("bob", "p1", domain.VerdictPending, 999, t1),
			},
			usernames: map[string]string{"alice": "alice", "bob": "bob"},
			counts: map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 2,
				submissionCountKey("bob", "p1"):   2,
			},
			wantOrder:  []string{"alice", "bob"},
			wantTotals: []int{0, 0},
			wantSolved: []int{0, 0},
			wantLast:   []*time.Time{nil, nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := services.BuildStandingsEntries(tt.submissions, tt.usernames, tt.counts)

			require.Len(t, entries, len(tt.wantOrder))
			assert.Equal(t, tt.wantOrder, standingsUserIDs(entries))
			for i, entry := range entries {
				assert.Equal(t, tt.wantTotals[i], entry.TotalScore, "entry %d total score", i)
				assert.Equal(t, tt.wantSolved[i], entry.SolvedCount, "entry %d solved count", i)
				if tt.wantLast[i] == nil {
					assert.Nil(t, entry.LastAcceptedAt, "entry %d last accepted", i)
				} else {
					require.NotNil(t, entry.LastAcceptedAt, "entry %d last accepted", i)
					assert.True(t, tt.wantLast[i].Equal(*entry.LastAcceptedAt), "entry %d last accepted", i)
				}
			}
		})
	}
}

func TestBuildStandingsEntriesProblemScores(t *testing.T) {
	base := standingsBaseTime()
	t1 := base
	t2 := base.Add(time.Hour)

	t.Run("unattempted problem contributes nothing and does not appear", func(t *testing.T) {
		entries := services.BuildStandingsEntries(
			[]domain.Submission{
				acceptedSubmission("alice", "p1", 100, t1),
				acceptedSubmission("bob", "p1", 100, t1),
				acceptedSubmission("bob", "p2", 50, t2),
			},
			map[string]string{"alice": "alice", "bob": "bob"},
			map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 1,
				submissionCountKey("bob", "p1"):   1,
				submissionCountKey("bob", "p2"):   1,
			},
		)

		require.Len(t, entries, 2)
		alice, ok := standingsEntryByUser(entries, "alice")
		require.True(t, ok)
		assert.Equal(t, 100, alice.TotalScore)
		_, found := standingsProblemScore(alice, "p2")
		assert.False(t, found, "unattempted problem must not appear in problem scores")

		p1, found := standingsProblemScore(alice, "p1")
		require.True(t, found)
		assert.Equal(t, 100, p1.BestScore)
		assert.Equal(t, 1, p1.Attempts)
	})

	t.Run("attempts come from counts and fall back to accepted count", func(t *testing.T) {
		entries := services.BuildStandingsEntries(
			[]domain.Submission{
				acceptedSubmission("alice", "p1", 50, t2),
				acceptedSubmission("alice", "p1", 100, t1),
				acceptedSubmission("bob", "p1", 100, t1),
			},
			map[string]string{"alice": "alice", "bob": "bob"},
			map[domain.SubmissionCountKey]int{
				submissionCountKey("bob", "p1"): 5,
			},
		)

		require.Len(t, entries, 2)
		alice, ok := standingsEntryByUser(entries, "alice")
		require.True(t, ok)
		p1, found := standingsProblemScore(alice, "p1")
		require.True(t, found)
		assert.Equal(t, 2, p1.Attempts, "no count entry: attempts fall back to accepted count")
		assert.Equal(t, 100, p1.BestScore)
		require.NotNil(t, p1.FirstAcceptedAt)
		assert.True(t, t1.Equal(*p1.FirstAcceptedAt))

		bob, ok := standingsEntryByUser(entries, "bob")
		require.True(t, ok)
		p1, found = standingsProblemScore(bob, "p1")
		require.True(t, found)
		assert.Equal(t, 5, p1.Attempts, "count entry includes failed attempts")
	})

	t.Run("failed-only problem score is zero with attempts and no first accepted time", func(t *testing.T) {
		entries := services.BuildStandingsEntries(
			[]domain.Submission{
				rejectedSubmission("alice", "p1", 0, t1),
			},
			map[string]string{"alice": "alice"},
			map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 3,
			},
		)

		require.Len(t, entries, 1)
		assert.Equal(t, 0, entries[0].TotalScore)
		assert.Equal(t, 0, entries[0].SolvedCount)
		assert.Nil(t, entries[0].LastAcceptedAt)
		p1, found := standingsProblemScore(entries[0], "p1")
		require.True(t, found)
		assert.Equal(t, 0, p1.BestScore)
		assert.Equal(t, 3, p1.Attempts)
		assert.Nil(t, p1.FirstAcceptedAt)
	})

	t.Run("problem scores are sorted by problem id", func(t *testing.T) {
		entries := services.BuildStandingsEntries(
			[]domain.Submission{
				acceptedSubmission("alice", "p2", 50, t2),
				acceptedSubmission("alice", "p1", 100, t1),
			},
			map[string]string{"alice": "alice"},
			map[domain.SubmissionCountKey]int{
				submissionCountKey("alice", "p1"): 1,
				submissionCountKey("alice", "p2"): 1,
			},
		)

		require.Len(t, entries, 1)
		require.Len(t, entries[0].ProblemScores, 2)
		assert.Equal(t, "p1", entries[0].ProblemScores[0].ProblemID)
		assert.Equal(t, "p2", entries[0].ProblemScores[1].ProblemID)
	})

	t.Run("username falls back to user id when unknown", func(t *testing.T) {
		entries := services.BuildStandingsEntries(
			[]domain.Submission{
				acceptedSubmission("user-9", "p1", 100, t1),
			},
			map[string]string{},
			map[domain.SubmissionCountKey]int{
				submissionCountKey("user-9", "p1"): 1,
			},
		)

		require.Len(t, entries, 1)
		assert.Equal(t, "user-9", entries[0].Username)
	})
}

func TestBuildStandingsEntriesCountsOnlyAttempts(t *testing.T) {
	entries := services.BuildStandingsEntries(
		nil,
		map[string]string{"carol": "carol"},
		map[domain.SubmissionCountKey]int{submissionCountKey("carol", "p1"): 2},
	)
	require.Len(t, entries, 1)
	require.Len(t, entries[0].ProblemScores, 1)
	assert.Equal(t, 0, entries[0].ProblemScores[0].BestScore)
	assert.Equal(t, 2, entries[0].ProblemScores[0].Attempts)
	assert.Nil(t, entries[0].ProblemScores[0].FirstAcceptedAt)
}

type fakeStandingsContestRepo struct {
	contest    *domain.Contest
	nilContest bool
	err        error
}

func (f *fakeStandingsContestRepo) CreateContest(ctx context.Context, contest *domain.Contest) error {
	return nil
}

func (f *fakeStandingsContestRepo) GetByID(ctx context.Context, id string) (*domain.Contest, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.nilContest {
		return nil, nil
	}
	return f.contest, nil
}

func (f *fakeStandingsContestRepo) UpdateContest(ctx context.Context, contest *domain.Contest) error {
	return nil
}

func (f *fakeStandingsContestRepo) DeleteContest(ctx context.Context, id string) error { return nil }

func (f *fakeStandingsContestRepo) CheckContestInTimeWindow(ctx context.Context, startTime, endTime time.Time) ([]domain.Contest, error) {
	return nil, nil
}

func (f *fakeStandingsContestRepo) GetAllContests(ctx context.Context) ([]domain.Contest, error) {
	return nil, nil
}

type fakeStandingsRepository struct {
	accepted    []domain.Submission
	acceptedErr error
	counts      map[domain.SubmissionCountKey]int
	countsErr   error
}

func (f *fakeStandingsRepository) GetAcceptedSubmissionsForContest(ctx context.Context, contestID string) ([]domain.Submission, error) {
	if f.acceptedErr != nil {
		return nil, f.acceptedErr
	}
	return f.accepted, nil
}

func (f *fakeStandingsRepository) GetSubmissionCountsForContest(ctx context.Context, contestID string) (map[domain.SubmissionCountKey]int, error) {
	if f.countsErr != nil {
		return nil, f.countsErr
	}
	return f.counts, nil
}

type fakeUsernameRepository struct {
	usernames map[string]string
	err       error
	calls     int
	gotIDs    [][]string
}

func (f *fakeUsernameRepository) GetUsernamesByIDs(ctx context.Context, ids []string) (map[string]string, error) {
	f.calls++
	f.gotIDs = append(f.gotIDs, append([]string(nil), ids...))
	if f.err != nil {
		return nil, f.err
	}
	if f.usernames == nil {
		return map[string]string{}, nil
	}
	return f.usernames, nil
}

func standingsContestAt(now time.Time, isActive bool, start, end time.Time) *domain.Contest {
	return &domain.Contest{
		Id:             "contest-1",
		Name:           "contest",
		StartTime:      start,
		EndTime:        end,
		IsActive:       isActive,
		CreatedBy:      "creator-1",
		ProblemSetters: []string{"setter-1"},
	}
}

func newStandingsUseCase(contestRepo *fakeStandingsContestRepo, standingsRepo *fakeStandingsRepository, usernameRepo *fakeUsernameRepository) domain.StandingsUseCase {
	return services.NewStandingsService(standingsRepo, usernameRepo, contestRepo)
}

func TestGetStandingsUpcomingAccess(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		role    string
		userID  string
		wantErr error
	}{
		{name: "admin may view upcoming standings", role: "admin", userID: "admin-1"},
		{name: "problem-setter role may view upcoming standings", role: "problem-setter", userID: "setter-9"},
		{name: "listed setter may view upcoming standings", role: "user", userID: "setter-1"},
		{name: "creator is not privileged before start", role: "user", userID: "creator-1", wantErr: domain.ErrStandingsForbidden},
		{name: "participant may not view upcoming standings", role: "user", userID: "participant-1", wantErr: domain.ErrStandingsForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contestRepo := &fakeStandingsContestRepo{contest: standingsContestAt(now, true, now.Add(time.Hour), now.Add(2*time.Hour))}
			standingsRepo := &fakeStandingsRepository{counts: map[domain.SubmissionCountKey]int{}}
			usernameRepo := &fakeUsernameRepository{}
			useCase := newStandingsUseCase(contestRepo, standingsRepo, usernameRepo)

			resp, err := useCase.GetStandings(context.Background(), "contest-1", tt.userID, tt.role)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, resp)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, domain.ContestStatusUpcoming, resp.Status)
			assert.Equal(t, "contest-1", resp.ContestID)
			require.NotNil(t, resp.Entries)
			assert.Empty(t, resp.Entries)
		})
	}
}

func TestGetStandingsLiveEndedAndInactiveInWindow(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name       string
		contest    *domain.Contest
		wantStatus string
	}{
		{
			name:       "live contest allows any caller",
			contest:    standingsContestAt(now, true, now.Add(-time.Hour), now.Add(time.Hour)),
			wantStatus: domain.ContestStatusLive,
		},
		{
			name:       "ended contest allows any caller",
			contest:    standingsContestAt(now, true, now.Add(-2*time.Hour), now.Add(-time.Hour)),
			wantStatus: domain.ContestStatusEnded,
		},
		{
			name:       "inactive contest inside its window is upcoming",
			contest:    standingsContestAt(now, false, now.Add(-time.Hour), now.Add(time.Hour)),
			wantStatus: domain.ContestStatusUpcoming,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contestRepo := &fakeStandingsContestRepo{contest: tt.contest}
			standingsRepo := &fakeStandingsRepository{counts: map[domain.SubmissionCountKey]int{}}
			usernameRepo := &fakeUsernameRepository{}
			useCase := newStandingsUseCase(contestRepo, standingsRepo, usernameRepo)

			resp, err := useCase.GetStandings(context.Background(), "contest-1", "participant-1", "user")
			if tt.wantStatus == domain.ContestStatusUpcoming {
				require.Error(t, err)
				assert.ErrorIs(t, err, domain.ErrStandingsForbidden)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.wantStatus, resp.Status)
		})
	}
}

func TestGetStandingsInactiveInWindowStatusForAdmin(t *testing.T) {
	now := time.Now()
	contestRepo := &fakeStandingsContestRepo{contest: standingsContestAt(now, false, now.Add(-time.Hour), now.Add(time.Hour))}
	useCase := newStandingsUseCase(contestRepo, &fakeStandingsRepository{counts: map[domain.SubmissionCountKey]int{}}, &fakeUsernameRepository{})

	resp, err := useCase.GetStandings(context.Background(), "contest-1", "admin-1", "admin")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, domain.ContestStatusUpcoming, resp.Status)
}

func TestGetStandingsMissingContest(t *testing.T) {
	tests := []struct {
		name        string
		contestRepo *fakeStandingsContestRepo
	}{
		{name: "record not found", contestRepo: &fakeStandingsContestRepo{err: gorm.ErrRecordNotFound}},
		{name: "nil contest without error", contestRepo: &fakeStandingsContestRepo{nilContest: true}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := newStandingsUseCase(tt.contestRepo, &fakeStandingsRepository{}, &fakeUsernameRepository{})
			resp, err := useCase.GetStandings(context.Background(), "contest-1", "user-1", "user")
			require.Error(t, err)
			assert.ErrorIs(t, err, domain.ErrContestNotFound)
			assert.Nil(t, resp)
		})
	}
}

func TestGetStandingsRepoErrors(t *testing.T) {
	now := time.Now()
	boom := errors.New("boom")
	live := standingsContestAt(now, true, now.Add(-time.Hour), now.Add(time.Hour))

	tests := []struct {
		name          string
		contestRepo   *fakeStandingsContestRepo
		standingsRepo *fakeStandingsRepository
		usernameRepo  *fakeUsernameRepository
	}{
		{
			name:          "contest repo error is wrapped",
			contestRepo:   &fakeStandingsContestRepo{err: boom},
			standingsRepo: &fakeStandingsRepository{},
			usernameRepo:  &fakeUsernameRepository{},
		},
		{
			name:          "accepted submissions error is wrapped",
			contestRepo:   &fakeStandingsContestRepo{contest: live},
			standingsRepo: &fakeStandingsRepository{acceptedErr: boom},
			usernameRepo:  &fakeUsernameRepository{},
		},
		{
			name:          "submission counts error is wrapped",
			contestRepo:   &fakeStandingsContestRepo{contest: live},
			standingsRepo: &fakeStandingsRepository{countsErr: boom},
			usernameRepo:  &fakeUsernameRepository{},
		},
		{
			name:          "username lookup failure fails the request",
			contestRepo:   &fakeStandingsContestRepo{contest: live},
			standingsRepo: &fakeStandingsRepository{accepted: []domain.Submission{acceptedSubmission("alice", "p1", 100, now)}, counts: map[domain.SubmissionCountKey]int{submissionCountKey("alice", "p1"): 1}},
			usernameRepo:  &fakeUsernameRepository{err: boom},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useCase := newStandingsUseCase(tt.contestRepo, tt.standingsRepo, tt.usernameRepo)
			resp, err := useCase.GetStandings(context.Background(), "contest-1", "user-1", "user")
			require.Error(t, err)
			assert.ErrorIs(t, err, boom)
			assert.Nil(t, resp)
		})
	}
}

func TestGetStandingsUsernameBatchLoadAndFallback(t *testing.T) {
	now := time.Now()
	live := standingsContestAt(now, true, now.Add(-time.Hour), now.Add(time.Hour))
	standingsRepo := &fakeStandingsRepository{
		accepted: []domain.Submission{
			acceptedSubmission("user-a", "p1", 100, now),
			acceptedSubmission("user-b", "p1", 50, now),
		},
		counts: map[domain.SubmissionCountKey]int{
			submissionCountKey("user-a", "p1"): 1,
			submissionCountKey("user-b", "p1"): 1,
			submissionCountKey("user-c", "p1"): 2,
		},
	}
	usernameRepo := &fakeUsernameRepository{usernames: map[string]string{"user-a": "alice"}}
	useCase := newStandingsUseCase(&fakeStandingsContestRepo{contest: live}, standingsRepo, usernameRepo)

	resp, err := useCase.GetStandings(context.Background(), "contest-1", "user-1", "user")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 1, usernameRepo.calls, "usernames must load in a single batch call")
	require.Len(t, usernameRepo.gotIDs, 1)
	assert.ElementsMatch(t, []string{"user-a", "user-b", "user-c"}, usernameRepo.gotIDs[0])

	byUser := map[string]domain.StandingsEntry{}
	for _, entry := range resp.Entries {
		byUser[entry.UserID] = entry
	}
	require.Len(t, byUser, 3)
	assert.Equal(t, "alice", byUser["user-a"].Username)
	assert.Equal(t, "user-b", byUser["user-b"].Username, "missing user falls back to user id")
	assert.Equal(t, "user-c", byUser["user-c"].Username, "counts-only user falls back to user id")
	assert.Equal(t, 0, byUser["user-c"].TotalScore)
}

func TestGetStandingsEmptyContestSkipsUsernameLookup(t *testing.T) {
	now := time.Now()
	live := standingsContestAt(now, true, now.Add(-time.Hour), now.Add(time.Hour))
	usernameRepo := &fakeUsernameRepository{}
	useCase := newStandingsUseCase(
		&fakeStandingsContestRepo{contest: live},
		&fakeStandingsRepository{counts: map[domain.SubmissionCountKey]int{}},
		usernameRepo,
	)

	resp, err := useCase.GetStandings(context.Background(), "contest-1", "user-1", "user")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Entries)
	assert.Equal(t, 0, usernameRepo.calls)
}

func TestGetStandingsNilRepos(t *testing.T) {
	useCase := services.NewStandingsService(nil, nil, nil)
	resp, err := useCase.GetStandings(context.Background(), "contest-1", "user-1", "user")
	require.Error(t, err)
	assert.Nil(t, resp)
}
