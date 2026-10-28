package services_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/services"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
