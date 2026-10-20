package domain_test

import (
	"algoforces/internal/domain"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUserCreation(t *testing.T) {
	tests := []struct {
		name    string
		user    domain.User
		isValid bool
	}{
		{
			name: "Valid user",
			user: domain.User{
				Id:       uuid.New().String(),
				Email:    "test@example.com",
				Password: "hashedpassword123",
				Role:     "user",
				Username: "testuser",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			isValid: true,
		},
		{
			name: "User with admin role",
			user: domain.User{
				Id:       uuid.New().String(),
				Email:    "admin@example.com",
				Password: "hashedpassword123",
				Role:     "admin",
				Username: "admin",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				if tt.user.Id == "" {
					t.Error("Valid user should have an ID")
				}
				if tt.user.Email == "" {
					t.Error("Valid user should have an email")
				}
			}
		})
	}
}

func TestProblemCreation(t *testing.T) {
	tests := []struct {
		name      string
		problem   domain.Problem
		isValid   bool
	}{
		{
			name: "Valid problem",
			problem: domain.Problem{
				UniqueID:           uuid.New().String(),
				Title:              "Two Sum",
				Statement:          "Find two numbers that add up to target",
				Difficulty:         "easy",
				TimeLimitInSeconds: 1,
				MemoryLimitInMB:    256,
				CreatedBy:          uuid.New().String(),
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				if tt.problem.UniqueID == "" {
					t.Error("Valid problem should have an ID")
				}
				if tt.problem.Title == "" {
					t.Error("Valid problem should have a title")
				}
			}
		})
	}
}

func TestTestCaseCreation(t *testing.T) {
	tests := []struct {
		name      string
		testcase  domain.TestCase
		isValid   bool
	}{
		{
			name: "Valid test case",
			testcase: domain.TestCase{
				UniqueID:       uuid.New().String(),
				ProblemID:      uuid.New().String(),
				Input:          "1 2",
				ExpectedOutput: "3",
				IsHidden:       false,
				OrderPosition:  1,
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				if tt.testcase.UniqueID == "" {
					t.Error("Valid test case should have an ID")
				}
				if tt.testcase.ProblemID == "" {
					t.Error("Valid test case should have a problem ID")
				}
			}
		})
	}
}

func TestContestCreation(t *testing.T) {
	startTime := time.Now()
	endTime := startTime.Add(2 * time.Hour)

	tests := []struct {
		name     string
		contest  domain.Contest
		isValid  bool
	}{
		{
			name: "Valid contest",
			contest: domain.Contest{
				Id:          uuid.New().String(),
				Name:        "Contest 1",
				Description: "First contest",
				StartTime:   startTime,
				EndTime:     endTime,
				Duration:    120,
				Visible:     true,
				IsActive:    true,
				CreatedBy:   uuid.New().String(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				if tt.contest.Id == "" {
					t.Error("Valid contest should have an ID")
				}
			}
		})
	}
}

func TestSubmissionCreation(t *testing.T) {
	tests := []struct {
		name       string
		submission domain.Submission
		isValid    bool
	}{
		{
			name: "Valid submission",
			submission: domain.Submission{
				UniqueID:       uuid.New().String(),
				UserId:         uuid.New().String(),
				ContestID:      uuid.New().String(),
				ProblemID:      uuid.New().String(),
				Code:           "print('hello')",
				Language:       "python",
				SubmittedAt:    time.Now(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				Verdict:        "Pending",
				Score:          0,
				TestCasesPassed: 0,
				TotalTestCases: 3,
			},
			isValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isValid {
				if tt.submission.UniqueID == "" {
					t.Error("Valid submission should have an ID")
				}
			}
		})
	}
}

func BenchmarkUserCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = domain.User{
			Id:        uuid.New().String(),
			Email:     "test@example.com",
			Password:  "hashedpassword",
			Role:      "user",
			Username:  "testuser",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
	}
}
