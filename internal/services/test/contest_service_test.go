package services_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/services"
	"context"
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

// fakeContestRepository is an in-memory domain.ContestRepository. Its
// overlap check mirrors the postgres SQL (start_time < end AND
// end_time > start, plus id <> excludeID when excludeID is non-empty).
type fakeContestRepository struct {
	contests []domain.Contest
}

func (f *fakeContestRepository) CreateContest(_ context.Context, contest *domain.Contest) error {
	f.contests = append(f.contests, *contest)
	return nil
}

func (f *fakeContestRepository) GetByID(_ context.Context, id string) (*domain.Contest, error) {
	for _, c := range f.contests {
		if c.Id == id {
			contest := c
			return &contest, nil
		}
	}
	return nil, errors.New("contest not found")
}

func (f *fakeContestRepository) UpdateContest(_ context.Context, contest *domain.Contest) error {
	for i, c := range f.contests {
		if c.Id == contest.Id {
			f.contests[i] = *contest
			return nil
		}
	}
	return errors.New("contest not found")
}

func (f *fakeContestRepository) DeleteContest(_ context.Context, id string) error {
	for i, c := range f.contests {
		if c.Id == id {
			f.contests = append(f.contests[:i], f.contests[i+1:]...)
			return nil
		}
	}
	return errors.New("contest not found")
}

func (f *fakeContestRepository) CheckContestInTimeWindow(ctx context.Context, startTime, endTime time.Time) ([]domain.Contest, error) {
	return f.CheckContestInTimeWindowExcluding(ctx, startTime, endTime, "")
}

func (f *fakeContestRepository) CheckContestInTimeWindowExcluding(_ context.Context, startTime, endTime time.Time, excludeID string) ([]domain.Contest, error) {
	var overlapping []domain.Contest
	for _, c := range f.contests {
		if excludeID != "" && c.Id == excludeID {
			continue
		}
		if c.StartTime.Before(endTime) && c.EndTime.After(startTime) {
			overlapping = append(overlapping, c)
		}
	}
	return overlapping, nil
}

func (f *fakeContestRepository) GetAllContests(_ context.Context) ([]domain.Contest, error) {
	contests := make([]domain.Contest, len(f.contests))
	copy(contests, f.contests)
	return contests, nil
}

func (f *fakeContestRepository) GetVisibleContests(_ context.Context) ([]domain.Contest, error) {
	var visible []domain.Contest
	for _, c := range f.contests {
		if c.Visible {
			visible = append(visible, c)
		}
	}
	sort.Slice(visible, func(i, j int) bool { return visible[i].StartTime.Before(visible[j].StartTime) })
	return visible, nil
}

// fakeUserRepository is an in-memory domain.UserRepository.
type fakeUserRepository struct {
	users map[string]*domain.User
}

func newFakeUserRepository(users ...*domain.User) *fakeUserRepository {
	f := &fakeUserRepository{users: map[string]*domain.User{}}
	for _, u := range users {
		f.users[u.Id] = u
	}
	return f
}

func (f *fakeUserRepository) Create(_ context.Context, user *domain.User) error {
	f.users[user.Id] = user
	return nil
}

func (f *fakeUserRepository) GetByEmail(_ context.Context, email string) (*domain.User, error) {
	for _, u := range f.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, errors.New("user not found")
}

func (f *fakeUserRepository) GetByID(_ context.Context, id string) (*domain.User, error) {
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return nil, errors.New("user not found")
}

func (f *fakeUserRepository) UpdateByID(_ context.Context, id string, user *domain.User) error {
	f.users[id] = user
	return nil
}

// fakeContestProblemsRepository is an in-memory
// domain.ContestProblemsRepository; only the by-contest lookup is used by
// the contest service, the rest are stubs.
type fakeContestProblemsRepository struct {
	problems []domain.ContestProblems
}

func (f *fakeContestProblemsRepository) CreateContestProblem(_ context.Context, contestProblem *domain.ContestProblems) error {
	f.problems = append(f.problems, *contestProblem)
	return nil
}

func (f *fakeContestProblemsRepository) GetContestProblemByID(_ context.Context, uniqueID string) (*domain.ContestProblems, error) {
	for _, p := range f.problems {
		if p.UniqueID == uniqueID {
			problem := p
			return &problem, nil
		}
	}
	return nil, errors.New("contest problem not found")
}

func (f *fakeContestProblemsRepository) GetContestProblemsByContestID(_ context.Context, contestID string) ([]domain.ContestProblems, error) {
	var problems []domain.ContestProblems
	for _, p := range f.problems {
		if p.ContestID == contestID {
			problems = append(problems, p)
		}
	}
	return problems, nil
}

func (f *fakeContestProblemsRepository) UpdateContestProblem(_ context.Context, contestProblem *domain.ContestProblems) error {
	for i, p := range f.problems {
		if p.UniqueID == contestProblem.UniqueID {
			f.problems[i] = *contestProblem
			return nil
		}
	}
	return errors.New("contest problem not found")
}

func (f *fakeContestProblemsRepository) DeleteContestProblem(_ context.Context, uniqueID string) error {
	for i, p := range f.problems {
		if p.UniqueID == uniqueID {
			f.problems = append(f.problems[:i], f.problems[i+1:]...)
			return nil
		}
	}
	return errors.New("contest problem not found")
}

func (f *fakeContestProblemsRepository) GetContestProblemsByContestIDWithDetails(_ context.Context, contestID string) ([]domain.ContestProblemDetail, error) {
	problems, _ := f.GetContestProblemsByContestID(context.Background(), contestID)
	var details []domain.ContestProblemDetail
	for _, p := range problems {
		details = append(details, domain.ContestProblemDetail{
			UniqueID:      p.UniqueID,
			ContestID:     p.ContestID,
			ProblemID:     p.ProblemID,
			OrderPosition: p.OrderPosition,
			MaxPoints:     p.MaxPoints,
			CreatedAt:     p.CreatedAt,
			UpdatedAt:     p.UpdatedAt,
		})
	}
	return details, nil
}

func newContestTestService() (domain.ContestUseCase, *fakeContestRepository, *fakeUserRepository, *fakeContestProblemsRepository) {
	contestRepo := &fakeContestRepository{}
	userRepo := newFakeUserRepository(
		&domain.User{Id: "setter-id", Email: "setter@example.com", Role: "problem-setter", Username: "setter"},
		&domain.User{Id: "admin-id", Email: "admin@example.com", Role: "admin", Username: "admin"},
	)
	problemsRepo := &fakeContestProblemsRepository{}
	svc := services.NewContestService(contestRepo, userRepo, problemsRepo)
	return svc, contestRepo, userRepo, problemsRepo
}

func TestFakeContestRepositoryOverlapExcludesSelf(t *testing.T) {
	ctx := context.Background()
	repo := &fakeContestRepository{}
	start := time.Date(2026, 10, 27, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 10, 27, 12, 0, 0, 0, time.UTC)
	if err := repo.CreateContest(ctx, &domain.Contest{Id: "c1", Name: "Contest 1", StartTime: start, EndTime: end}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}

	overlappingStart := time.Date(2026, 10, 27, 11, 0, 0, 0, time.UTC)
	overlappingEnd := time.Date(2026, 10, 27, 13, 0, 0, 0, time.UTC)

	found, err := repo.CheckContestInTimeWindow(ctx, overlappingStart, overlappingEnd)
	if err != nil {
		t.Fatalf("CheckContestInTimeWindow() error = %v", err)
	}
	if len(found) != 1 {
		t.Errorf("CheckContestInTimeWindow() returned %d contests, want 1", len(found))
	}

	excluded, err := repo.CheckContestInTimeWindowExcluding(ctx, overlappingStart, overlappingEnd, "c1")
	if err != nil {
		t.Fatalf("CheckContestInTimeWindowExcluding() error = %v", err)
	}
	if len(excluded) != 0 {
		t.Errorf("CheckContestInTimeWindowExcluding(self) returned %d contests, want 0", len(excluded))
	}

	otherExcluded, err := repo.CheckContestInTimeWindowExcluding(ctx, overlappingStart, overlappingEnd, "other-id")
	if err != nil {
		t.Fatalf("CheckContestInTimeWindowExcluding() error = %v", err)
	}
	if len(otherExcluded) != 1 {
		t.Errorf("CheckContestInTimeWindowExcluding(other) returned %d contests, want 1", len(otherExcluded))
	}

	// Windows that only touch a boundary do not overlap (strict < / >).
	touchingEnd, _ := repo.CheckContestInTimeWindow(ctx, end, end.Add(time.Hour))
	if len(touchingEnd) != 0 {
		t.Errorf("window starting at existing end returned %d contests, want 0", len(touchingEnd))
	}
	touchingStart, _ := repo.CheckContestInTimeWindow(ctx, start.Add(-time.Hour), start)
	if len(touchingStart) != 0 {
		t.Errorf("window ending at existing start returned %d contests, want 0", len(touchingStart))
	}
}

func TestContestServiceCreateContestOverlap(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()
	now := time.Now()
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: uuid.New().String(), Name: "Existing",
		StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour),
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}

	_, err := svc.CreateContest(ctx, &domain.CreateContestRequest{
		Name: "Overlapping", StartTime: now.Add(3 * time.Hour), EndTime: now.Add(5 * time.Hour), Duration: 120,
	}, "admin-id")
	if err == nil || !strings.Contains(err.Error(), "already a contest exists") {
		t.Errorf("CreateContest() overlapping error = %v, want time-window error", err)
	}

	created, err := svc.CreateContest(ctx, &domain.CreateContestRequest{
		Name: "Later", StartTime: now.Add(5 * time.Hour), EndTime: now.Add(6 * time.Hour), Duration: 60,
	}, "admin-id")
	if err != nil {
		t.Fatalf("CreateContest() non-overlapping error = %v", err)
	}
	if created.Status != "upcoming" {
		t.Errorf("CreateContest() status = %q, want %q", created.Status, "upcoming")
	}
}

func TestContestServiceUpdateContestExcludesSelf(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()
	now := time.Now()
	contestID := uuid.New().String()
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: contestID, Name: "Old name",
		StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour),
		Duration: 120, Visible: true, ProblemSetters: []string{"setter-id"},
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}

	// Updating with the same window must not collide with itself.
	updated, err := svc.UpdateContest(ctx, &domain.UpdateContestRequest{
		Id: contestID, Name: "New name",
		StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour),
		Duration: 120, Visible: true, ProblemSetters: []string{"setter-id"},
	})
	if err != nil {
		t.Fatalf("UpdateContest() error = %v, want success (self must be excluded)", err)
	}
	if updated.Name != "New name" {
		t.Errorf("UpdateContest() name = %q, want %q", updated.Name, "New name")
	}
	if updated.Status != "upcoming" {
		t.Errorf("UpdateContest() status = %q, want %q", updated.Status, "upcoming")
	}
}

func TestContestServiceUpdateContestRejectsOverlappingOther(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()
	now := time.Now()
	contestAID := uuid.New().String()
	contestBStart := now.Add(6 * time.Hour)
	contestBEnd := now.Add(8 * time.Hour)
	for _, c := range []domain.Contest{
		{Id: contestAID, Name: "A", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour), Duration: 120},
		{Id: uuid.New().String(), Name: "B", StartTime: contestBStart, EndTime: contestBEnd, Duration: 120},
	} {
		if err := contestRepo.CreateContest(ctx, &c); err != nil {
			t.Fatalf("CreateContest() error = %v", err)
		}
	}

	_, err := svc.UpdateContest(ctx, &domain.UpdateContestRequest{
		Id: contestAID, Name: "A moved",
		StartTime: now.Add(7 * time.Hour), EndTime: now.Add(9 * time.Hour),
		Duration: 120,
	})
	if err == nil || !strings.Contains(err.Error(), "already a contest exists") {
		t.Errorf("UpdateContest() overlapping-other error = %v, want time-window error", err)
	}
}

func TestContestServiceUpdateContestRejectsTimeChangeAfterStart(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()
	now := time.Now()
	contestID := uuid.New().String()
	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: contestID, Name: "Started",
		StartTime: start, EndTime: end, Duration: 120, IsActive: true,
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}

	_, err := svc.UpdateContest(ctx, &domain.UpdateContestRequest{
		Id: contestID, Name: "Started",
		StartTime: now.Add(-30 * time.Minute), EndTime: end, Duration: 120, IsActive: true,
	})
	if err == nil || !strings.Contains(err.Error(), "cannot change start time") {
		t.Errorf("UpdateContest() changed-start error = %v, want time-lock error", err)
	}

	// Renaming with unchanged times is still allowed after the start.
	updated, err := svc.UpdateContest(ctx, &domain.UpdateContestRequest{
		Id: contestID, Name: "Renamed",
		StartTime: start, EndTime: end, Duration: 120, IsActive: true,
	})
	if err != nil {
		t.Fatalf("UpdateContest() unchanged-times error = %v", err)
	}
	if updated.Name != "Renamed" {
		t.Errorf("UpdateContest() name = %q, want %q", updated.Name, "Renamed")
	}
}

func TestContestServiceStartContestRequiresProblem(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, problemsRepo := newContestTestService()
	now := time.Now()
	contestID := uuid.New().String()
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: contestID, Name: "No problems",
		StartTime: now.Add(time.Hour), EndTime: now.Add(3 * time.Hour), Duration: 120,
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}

	if err := svc.StartContest(ctx, contestID); err == nil || !strings.Contains(err.Error(), "at least one problem") {
		t.Errorf("StartContest() without problems error = %v, want problem-required error", err)
	}

	if err := problemsRepo.CreateContestProblem(ctx, &domain.ContestProblems{
		UniqueID: uuid.New().String(), ContestID: contestID, ProblemID: uuid.New().String(), OrderPosition: 1, MaxPoints: 100,
	}); err != nil {
		t.Fatalf("CreateContestProblem() error = %v", err)
	}
	if err := svc.StartContest(ctx, contestID); err != nil {
		t.Fatalf("StartContest() with a problem error = %v", err)
	}
	stored, err := contestRepo.GetByID(ctx, contestID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if !stored.IsActive {
		t.Error("StartContest() did not set IsActive=true")
	}
}

func TestContestServiceStartContestRejectsEnded(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, problemsRepo := newContestTestService()
	now := time.Now()
	contestID := uuid.New().String()
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: contestID, Name: "Past",
		StartTime: now.Add(-3 * time.Hour), EndTime: now.Add(-time.Hour), Duration: 120,
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}
	if err := problemsRepo.CreateContestProblem(ctx, &domain.ContestProblems{
		UniqueID: uuid.New().String(), ContestID: contestID, ProblemID: uuid.New().String(), OrderPosition: 1, MaxPoints: 100,
	}); err != nil {
		t.Fatalf("CreateContestProblem() error = %v", err)
	}

	if err := svc.StartContest(ctx, contestID); err == nil || !strings.Contains(err.Error(), "has ended") {
		t.Errorf("StartContest() ended error = %v, want ended error", err)
	}
}

func TestContestServiceEndContestSetsInactive(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()
	now := time.Now()
	contestID := uuid.New().String()
	end := now.Add(time.Hour)
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: contestID, Name: "Live",
		StartTime: now.Add(-time.Hour), EndTime: end, Duration: 120, IsActive: true,
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}

	if err := svc.EndContest(ctx, contestID); err != nil {
		t.Fatalf("EndContest() error = %v", err)
	}
	stored, err := contestRepo.GetByID(ctx, contestID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if stored.IsActive {
		t.Error("EndContest() did not set IsActive=false")
	}
	if !stored.EndTime.Equal(end) {
		t.Errorf("EndContest() changed EndTime to %v, want %v", stored.EndTime, end)
	}
	if got := domain.ContestStatus(*stored, time.Now()); got != "ended" {
		t.Errorf("ContestStatus() after EndContest = %q, want %q", got, "ended")
	}
}

func TestContestServiceDeleteContestRejectsLive(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()
	now := time.Now()
	liveID := uuid.New().String()
	upcomingID := uuid.New().String()
	for _, c := range []domain.Contest{
		{Id: liveID, Name: "Live", StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour), Duration: 120, IsActive: true},
		{Id: upcomingID, Name: "Upcoming", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour), Duration: 120},
	} {
		if err := contestRepo.CreateContest(ctx, &c); err != nil {
			t.Fatalf("CreateContest() error = %v", err)
		}
	}

	if err := svc.DeleteContest(ctx, &domain.DeleteContestRequest{Id: liveID}); err == nil || !strings.Contains(err.Error(), "live") {
		t.Errorf("DeleteContest() live error = %v, want live-contest error", err)
	}
	if err := svc.DeleteContest(ctx, &domain.DeleteContestRequest{Id: upcomingID}); err != nil {
		t.Errorf("DeleteContest() upcoming error = %v, want success", err)
	}
	if _, err := contestRepo.GetByID(ctx, upcomingID); err == nil {
		t.Error("DeleteContest() upcoming contest still exists")
	}
}

func TestContestServiceGetVisibleContests(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()
	now := time.Now()
	lateVisibleID := uuid.New().String()
	earlyVisibleID := uuid.New().String()
	for _, c := range []domain.Contest{
		{Id: lateVisibleID, Name: "Late visible", StartTime: now.Add(4 * time.Hour), EndTime: now.Add(6 * time.Hour), Duration: 120, Visible: true},
		{Id: uuid.New().String(), Name: "Hidden", StartTime: now.Add(time.Hour), EndTime: now.Add(2 * time.Hour), Duration: 60, Visible: false},
		{Id: earlyVisibleID, Name: "Early visible", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(3 * time.Hour), Duration: 60, Visible: true},
	} {
		if err := contestRepo.CreateContest(ctx, &c); err != nil {
			t.Fatalf("CreateContest() error = %v", err)
		}
	}

	contests, err := svc.GetVisibleContests(ctx)
	if err != nil {
		t.Fatalf("GetVisibleContests() error = %v", err)
	}
	if len(contests) != 2 {
		t.Fatalf("GetVisibleContests() returned %d contests, want 2 visible", len(contests))
	}
	if contests[0].Id != earlyVisibleID || contests[1].Id != lateVisibleID {
		t.Errorf("GetVisibleContests() order = [%s %s], want [%s %s] (StartTime ascending)",
			contests[0].Id, contests[1].Id, earlyVisibleID, lateVisibleID)
	}
}
