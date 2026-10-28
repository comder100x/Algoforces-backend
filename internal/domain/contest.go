package domain

import (
	"context"
	"time"

	"github.com/lib/pq"
)

type Contest struct {
	Id             string         `json:"id" gorm:"primaryKey";type:uuid"`
	Name           string         `json:"name" gorm:"not null"`
	Description    string         `json:"description" gorm:""`
	StartTime      time.Time      `json:"start_time" gorm:"not null"`
	EndTime        time.Time      `json:"end_time" gorm:"not null"`
	Duration       int            `json:"duration" gorm:"not null"` // in minutes
	Visible        bool           `json:"visible" gorm:"default:false"`
	CreatedBy      string         `json:"created_by" gorm:"type:uuid;not null"` //refrences User(Id)
	IsActive       bool           `json:"is_active" gorm:"default:false"`
	ProblemSetters pq.StringArray `json:"problem-setters" gorm:"type:text[]"`
	CreatedAt      time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
}

// ContestStatus is the derived lifecycle status of a contest. It is a
// string-based enum so its JSON rendering is unchanged ("upcoming",
// "live", "ended") while callers compare typed constants instead of
// raw string literals. There is no status column in the database.
type ContestStatus string

const (
	ContestStatusUpcoming ContestStatus = "upcoming"
	ContestStatusLive     ContestStatus = "live"
	ContestStatusEnded    ContestStatus = "ended"
)

// DeriveContestStatus derives the lifecycle status of a contest at a given moment.
//
// It is a pure function of the stored fields and the supplied time; there is
// no status column in the database and callers must pass the time the status
// is evaluated at (services pass time.Now()).
//
// The rules, evaluated in order, are:
//
//  1. ContestStatusUpcoming — now is before StartTime, regardless of IsActive. A
//     contest activated early still reads as upcoming until it starts.
//  2. ContestStatusEnded — now is after EndTime, regardless of IsActive.
//  3. ContestStatusLive — StartTime <= now <= EndTime and IsActive is true.
//  4. ContestStatusEnded — otherwise: the start time has been reached
//     (StartTime <= now <= EndTime) but IsActive is false, i.e. the contest
//     was ended (EndContest sets IsActive=false without moving EndTime) or
//     was never started. A started-but-inactive contest therefore reports
//     as ended.
//
// Boundary behaviour: now == StartTime with IsActive=true is "live", and
// now == EndTime with IsActive=true is "live"; with IsActive=false both
// boundaries are "ended".
//
// NOTE: the helper is named DeriveContestStatus, not ContestStatus:
// Go does not allow a type and a function to share the name
// ContestStatus, and the type name is the one the review asked for.
func DeriveContestStatus(contest Contest, now time.Time) ContestStatus {
	if now.Before(contest.StartTime) {
		return ContestStatusUpcoming
	}
	if now.After(contest.EndTime) {
		return ContestStatusEnded
	}
	if contest.IsActive {
		return ContestStatusLive
	}
	return ContestStatusEnded
}

type CreateContestRequest struct {
	Name           string    `json:"name" binding:"required"`
	Description    string    `json:"description"`
	StartTime      time.Time `json:"start_time" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required,gtfield=StartTime"`
	Duration       int       `json:"duration" binding:"required,gt=0"` // in minutes
	Visible        bool      `json:"visible"`
	IsActive       bool      `json:"is_active"`
	ProblemSetters []string  `json:"problem-setters"`
}

type CreateContestResponse struct {
	Id             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	Duration       int           `json:"duration"` // in minutes
	Visible        bool          `json:"visible"`
	IsActive       bool          `json:"is_active"`
	Status         ContestStatus `json:"status"`
	ProblemSetters []string      `json:"problem-setters"`
	CreatedBy      string        `json:"created_by"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type UpdateContestRequest struct {
	Id             string    `json:"id" binding:"required,uuid"`
	Name           string    `json:"name" binding:"required"`
	Description    string    `json:"description"`
	StartTime      time.Time `json:"start_time" binding:"required"`
	EndTime        time.Time `json:"end_time" binding:"required,gtfield=StartTime"`
	Duration       int       `json:"duration" binding:"required,gt=0"` // in minutes
	Visible        bool      `json:"visible"`
	ProblemSetters []string  `json:"problem-setters"`
	IsActive       bool      `json:"is_active"`
}

type UpdateContestResponse struct {
	Id             string        `json:"id"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	StartTime      time.Time     `json:"start_time"`
	EndTime        time.Time     `json:"end_time"`
	Duration       int           `json:"duration"` // in minutes
	Visible        bool          `json:"visible"`
	ProblemSetters []string      `json:"problem-setters"`
	IsActive       bool          `json:"is_active"`
	Status         ContestStatus `json:"status"`
	CreatedBy      string        `json:"created_by"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

type DeleteContestRequest struct {
	Id string `json:"id" binding:"required,uuid"`
}

type ContestRepository interface {
	CreateContest(ctx context.Context, contest *Contest) error
	GetByID(ctx context.Context, id string) (*Contest, error)
	UpdateContest(ctx context.Context, contest *Contest) error
	DeleteContest(ctx context.Context, id string) error
	CheckContestInTimeWindow(ctx context.Context, startTime, endTime time.Time) ([]Contest, error)
	CheckContestInTimeWindowExcluding(ctx context.Context, startTime, endTime time.Time, excludeID string) ([]Contest, error)
	GetAllContests(ctx context.Context) ([]Contest, error)
	GetVisibleContests(ctx context.Context) ([]Contest, error)
}

type ContestUseCase interface {
	CreateContest(ctx context.Context, req *CreateContestRequest, userId string) (*CreateContestResponse, error)
	UpdateContest(ctx context.Context, req *UpdateContestRequest) (*UpdateContestResponse, error)
	GetContestDetails(ctx context.Context, id string) (*CreateContestResponse, error)
	DeleteContest(ctx context.Context, req *DeleteContestRequest) error
	StartContest(ctx context.Context, id string) error
	EndContest(ctx context.Context, id string) error
	GetAllContests(ctx context.Context) ([]CreateContestResponse, error)
	GetVisibleContests(ctx context.Context) ([]CreateContestResponse, error)
}
