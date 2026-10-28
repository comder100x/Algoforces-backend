package services_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/services"
	"testing"
)

func TestCanViewHiddenTestData(t *testing.T) {
	tests := []struct {
		name string
		role string
		want bool
	}{
		{name: "admin may view hidden test data", role: "admin", want: true},
		{name: "problem-setter may view hidden test data", role: "problem-setter", want: true},
		{name: "regular user may not view hidden test data", role: "user", want: false},
		{name: "empty role may not view hidden test data", role: "", want: false},
		{name: "unknown role may not view hidden test data", role: "superuser", want: false},
		{name: "role matching is case-sensitive", role: "Admin", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := services.CanViewHiddenTestData(tt.role); got != tt.want {
				t.Errorf("CanViewHiddenTestData(%q) = %v, want %v", tt.role, got, tt.want)
			}
		})
	}
}

func TestVisibleContestSubmissions(t *testing.T) {
	ownID := "c3d4e5f6-3333-4333-8333-3456789abcde"
	otherID := "a1b2c3d4-1111-4111-8111-123456789abc"
	submissions := []domain.Submission{
		{UniqueID: "sub-newest", UserId: otherID},
		{UniqueID: "sub-middle", UserId: ownID},
		{UniqueID: "sub-oldest", UserId: otherID},
	}

	submissionIDs := func(list []domain.Submission) []string {
		ids := make([]string, 0, len(list))
		for _, submission := range list {
			ids = append(ids, submission.UniqueID)
		}
		return ids
	}

	tests := []struct {
		name   string
		role   string
		userID string
		want   []string
	}{
		{
			name:   "regular user sees only own submissions in repository order",
			role:   "user",
			userID: ownID,
			want:   []string{"sub-middle"},
		},
		{
			name:   "regular user with no submissions sees none",
			role:   "user",
			userID: "d4e5f6a7-4444-4444-8444-456789abcdef",
			want:   []string{},
		},
		{
			name:   "admin sees all submissions in repository order",
			role:   "admin",
			userID: ownID,
			want:   []string{"sub-newest", "sub-middle", "sub-oldest"},
		},
		{
			name:   "problem-setter sees all submissions in repository order",
			role:   "problem-setter",
			userID: ownID,
			want:   []string{"sub-newest", "sub-middle", "sub-oldest"},
		},
		{
			name:   "unknown role sees only own submissions",
			role:   "superuser",
			userID: ownID,
			want:   []string{"sub-middle"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := submissionIDs(services.VisibleContestSubmissions(submissions, tt.userID, tt.role))
			if len(got) != len(tt.want) {
				t.Fatalf("VisibleContestSubmissions() returned %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("VisibleContestSubmissions()[%d] = %q, want %q (full result %v)", i, got[i], tt.want[i], got)
				}
			}
		})
	}
}

func TestVisibleContestSubmissionsEmpty(t *testing.T) {
	got := services.VisibleContestSubmissions(nil, "some-user", "user")
	if len(got) != 0 {
		t.Errorf("VisibleContestSubmissions(nil) returned %d submissions, want 0", len(got))
	}
}
