package app

import (
	"algoforces/internal/conf"
	"algoforces/internal/domain"
	"algoforces/internal/utils"
	"algoforces/pkg/database"
	"strings"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm/clause"
)

const (
	seedAdminID         = "a1b2c3d4-1111-4111-8111-123456789abc"
	seedProblemSetterID = "b2c3d4e5-2222-4222-8222-23456789abcd"
	seedUserID          = "c3d4e5f6-3333-4333-8333-3456789abcde"
	seedContestID       = "d4e5f6a7-4444-4444-8444-456789abcdef"
	seedProblemOneID    = "e5f6a7b8-5555-4555-8555-56789abcdef1"
	seedProblemTwoID    = "f6a7b8c9-6666-4666-8666-6789abcdef12"
)

func shouldSeed() bool {
	env := strings.ToLower(conf.ENV)
	return env != "prod" && env != "production"
}

func seed(db *database.Database) error {
	password, err := utils.HashPassword("password123")
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	contestStart := now.Add(24 * time.Hour)
	contestEnd := contestStart.Add(2 * time.Hour)

	users := []domain.User{
		{Id: seedAdminID, Username: "Admin", Email: "admin@algoforces.local", Password: password, Role: "admin"},
		{Id: seedProblemSetterID, Username: "Problem Setter", Email: "setter@algoforces.local", Password: password, Role: "problem-setter"},
		{Id: seedUserID, Username: "Demo User", Email: "user@algoforces.local", Password: password, Role: "user"},
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&users).Error; err != nil {
		return err
	}

	problems := []domain.Problem{
		{
			UniqueID:           seedProblemOneID,
			Title:              "A Plus B",
			Statement:          "Given two integers a and b, print their sum.",
			Difficulty:         "easy",
			TimeLimitInSeconds: 1,
			MemoryLimitInMB:    256,
			CreatedBy:          seedProblemSetterID,
		},
		{
			UniqueID:           seedProblemTwoID,
			Title:              "Maximum of Three",
			Statement:          "Given three integers, print the maximum value.",
			Difficulty:         "easy",
			TimeLimitInSeconds: 1,
			MemoryLimitInMB:    256,
			CreatedBy:          seedProblemSetterID,
		},
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&problems).Error; err != nil {
		return err
	}

	testCases := []domain.TestCase{
		{UniqueID: "7a7b7c7d-7777-4777-8777-777777777771", ProblemID: seedProblemOneID, Input: "2 3", ExpectedOutput: "5", IsHidden: false, OrderPosition: 1},
		{UniqueID: "7b7c7d7e-7777-4777-8777-777777777772", ProblemID: seedProblemOneID, Input: "10 -4", ExpectedOutput: "6", IsHidden: true, OrderPosition: 2},
		{UniqueID: "7c7d7e7f-7777-4777-8777-777777777773", ProblemID: seedProblemTwoID, Input: "1 7 3", ExpectedOutput: "7", IsHidden: false, OrderPosition: 1},
		{UniqueID: "7d7e7f8a-7777-4777-8777-777777777774", ProblemID: seedProblemTwoID, Input: "-1 -5 -3", ExpectedOutput: "-1", IsHidden: true, OrderPosition: 2},
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&testCases).Error; err != nil {
		return err
	}

	contest := domain.Contest{
		Id:             seedContestID,
		Name:           "Algoforces Demo Contest",
		Description:    "Starter contest created automatically for local development.",
		StartTime:      contestStart,
		EndTime:        contestEnd,
		Duration:       120,
		Visible:        true,
		CreatedBy:      seedAdminID,
		IsActive:       true,
		ProblemSetters: pq.StringArray{seedProblemSetterID},
	}
	if err := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&contest).Error; err != nil {
		return err
	}

	contestProblems := []domain.ContestProblems{
		{UniqueID: "8a8b8c8d-8888-4888-8888-888888888881", ContestID: seedContestID, ProblemID: seedProblemOneID, OrderPosition: 1, MaxPoints: 100},
		{UniqueID: "8b8c8d8e-8888-4888-8888-888888888882", ContestID: seedContestID, ProblemID: seedProblemTwoID, OrderPosition: 2, MaxPoints: 100},
	}
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&contestProblems).Error
}
