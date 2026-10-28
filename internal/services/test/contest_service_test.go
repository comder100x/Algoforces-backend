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
	"gorm.io/gorm"
)

// fakeContestRepository is an in-memory domain.ContestRepository. Its
// overlap check mirrors the postgres SQL (start_time < end AND
// end_time > start, plus id <> excludeID when excludeID is non-empty).
type fakeContestRepository struct {
	contests      []domain.Contest
	checkErr      error
	createErr     error
	getErr        error
	updateErr     error
	deleteErr     error
	getAllErr     error
	getVisibleErr error
}

func (f *fakeContestRepository) CreateContest(_ context.Context, contest *domain.Contest) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.contests = append(f.contests, *contest)
	return nil
}

func (f *fakeContestRepository) GetByID(_ context.Context, id string) (*domain.Contest, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	for _, c := range f.contests {
		if c.Id == id {
			contest := c
			return &contest, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeContestRepository) UpdateContest(_ context.Context, contest *domain.Contest) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	for i, c := range f.contests {
		if c.Id == contest.Id {
			f.contests[i] = *contest
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakeContestRepository) DeleteContest(_ context.Context, id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	for i, c := range f.contests {
		if c.Id == id {
			f.contests = append(f.contests[:i], f.contests[i+1:]...)
			return nil
		}
	}
	return gorm.ErrRecordNotFound
}

func (f *fakeContestRepository) CheckContestInTimeWindow(ctx context.Context, startTime, endTime time.Time) ([]domain.Contest, error) {
	return f.CheckContestInTimeWindowExcluding(ctx, startTime, endTime, "")
}

func (f *fakeContestRepository) CheckContestInTimeWindowExcluding(_ context.Context, startTime, endTime time.Time, excludeID string) ([]domain.Contest, error) {
	if f.checkErr != nil {
		return nil, f.checkErr
	}
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
	if f.getAllErr != nil {
		return nil, f.getAllErr
	}
	contests := make([]domain.Contest, len(f.contests))
	copy(contests, f.contests)
	return contests, nil
}

func (f *fakeContestRepository) GetVisibleContests(_ context.Context) ([]domain.Contest, error) {
	if f.getVisibleErr != nil {
		return nil, f.getVisibleErr
	}
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
	users      map[string]*domain.User
	getByIDErr error
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
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeUserRepository) GetByID(_ context.Context, id string) (*domain.User, error) {
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}
	if u, ok := f.users[id]; ok {
		return u, nil
	}
	return nil, gorm.ErrRecordNotFound
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
	err      error
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
	if f.err != nil {
		return nil, f.err
	}
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
	if created.Status != domain.ContestStatusUpcoming {
		t.Errorf("CreateContest() status = %q, want %q", created.Status, domain.ContestStatusUpcoming)
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
	if updated.Status != domain.ContestStatusUpcoming {
		t.Errorf("UpdateContest() status = %q, want %q", updated.Status, domain.ContestStatusUpcoming)
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
	if got := domain.DeriveContestStatus(*stored, time.Now()); got != domain.ContestStatusEnded {
		t.Errorf("DeriveContestStatus() after EndContest = %q, want %q", got, domain.ContestStatusEnded)
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

func TestContestServiceGetContestDetails(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()
	now := time.Now()
	contestID := uuid.New().String()
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: contestID, Name: "Details",
		StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour),
		Duration: 120, Visible: true,
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}

	got, err := svc.GetContestDetails(ctx, contestID)
	if err != nil {
		t.Fatalf("GetContestDetails() error = %v", err)
	}
	if got.Id != contestID || got.Name != "Details" || got.Status != domain.ContestStatusUpcoming {
		t.Errorf("GetContestDetails() = %+v, want id %q name %q status upcoming", got, contestID, "Details")
	}

	_, err = svc.GetContestDetails(ctx, uuid.New().String())
	if !errors.Is(err, domain.ErrContestNotFound) {
		t.Errorf("GetContestDetails() missing error = %v, want errors.Is ErrContestNotFound", err)
	}

	contestRepo.getErr = errors.New("db down")
	_, err = svc.GetContestDetails(ctx, contestID)
	if errors.Is(err, domain.ErrContestNotFound) || err == nil || !strings.Contains(err.Error(), "failed to get contest") {
		t.Errorf("GetContestDetails() repo error = %v, want wrapped non-sentinel error", err)
	}
}

func TestContestServiceListContestsEmptyAndErrors(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()

	all, err := svc.GetAllContests(ctx)
	if err != nil {
		t.Fatalf("GetAllContests() error = %v", err)
	}
	if all == nil {
		t.Error("GetAllContests() on empty store returned nil, want non-nil empty slice (JSON [])")
	}
	visible, err := svc.GetVisibleContests(ctx)
	if err != nil {
		t.Fatalf("GetVisibleContests() error = %v", err)
	}
	if visible == nil {
		t.Error("GetVisibleContests() on empty store returned nil, want non-nil empty slice (JSON [])")
	}

	contestRepo.getAllErr = errors.New("db down")
	if _, err := svc.GetAllContests(ctx); err == nil || !strings.Contains(err.Error(), "failed to get contests") {
		t.Errorf("GetAllContests() repo error = %v, want wrapped error", err)
	}
	contestRepo.getVisibleErr = errors.New("db down")
	if _, err := svc.GetVisibleContests(ctx); err == nil || !strings.Contains(err.Error(), "failed to get contests") {
		t.Errorf("GetVisibleContests() repo error = %v, want wrapped error", err)
	}
}

func TestContestServiceCreateContestRepoErrors(t *testing.T) {
	ctx := context.Background()
	now := time.Now()
	req := func() *domain.CreateContestRequest {
		return &domain.CreateContestRequest{
			Name: "New", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour), Duration: 120,
		}
	}

	svc, contestRepo, userRepo, _ := newContestTestService()
	contestRepo.checkErr = errors.New("db down")
	if _, err := svc.CreateContest(ctx, req(), "admin-id"); err == nil || !strings.Contains(err.Error(), "failed to check contest time window") {
		t.Errorf("CreateContest() window error = %v, want wrapped error", err)
	}
	contestRepo.checkErr = nil

	contestRepo.createErr = errors.New("db down")
	if _, err := svc.CreateContest(ctx, req(), "admin-id"); err == nil || !strings.Contains(err.Error(), "failed to create contest") {
		t.Errorf("CreateContest() create error = %v, want wrapped error", err)
	}
	contestRepo.createErr = nil

	_, err := svc.CreateContest(ctx, &domain.CreateContestRequest{
		Name: "Ghost setter", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour),
		Duration: 120, ProblemSetters: []string{"no-such-user"},
	}, "admin-id")
	if !errors.Is(err, domain.ErrProblemSetterNotFound) {
		t.Errorf("CreateContest() unknown setter error = %v, want errors.Is ErrProblemSetterNotFound", err)
	}

	userRepo.getByIDErr = errors.New("db down")
	_, err = svc.CreateContest(ctx, &domain.CreateContestRequest{
		Name: "Setter lookup fails", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour),
		Duration: 120, ProblemSetters: []string{"setter-id"},
	}, "admin-id")
	if err == nil || errors.Is(err, domain.ErrProblemSetterNotFound) || !strings.Contains(err.Error(), "failed to get problem setter user") {
		t.Errorf("CreateContest() setter lookup error = %v, want wrapped non-sentinel error", err)
	}
	userRepo.getByIDErr = nil

	regular := &domain.User{Id: "regular-id", Email: "regular@example.com", Role: "user", Username: "regular"}
	if err := userRepo.Create(ctx, regular); err != nil {
		t.Fatalf("Create() user error = %v", err)
	}
	_, err = svc.CreateContest(ctx, &domain.CreateContestRequest{
		Name: "Regular setter", StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour),
		Duration: 120, ProblemSetters: []string{"regular-id"},
	}, "admin-id")
	if !errors.Is(err, domain.ErrProblemSetterNotAuthorized) {
		t.Errorf("CreateContest() wrong-role setter error = %v, want errors.Is ErrProblemSetterNotAuthorized", err)
	}
}

func TestContestServiceUpdateContestErrors(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()
	now := time.Now()

	_, err := svc.UpdateContest(ctx, &domain.UpdateContestRequest{
		Id: uuid.New().String(), Name: "Missing",
		StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour), Duration: 120,
	})
	if !errors.Is(err, domain.ErrContestNotFound) {
		t.Errorf("UpdateContest() missing error = %v, want errors.Is ErrContestNotFound", err)
	}

	contestID := uuid.New().String()
	start := now.Add(-time.Hour)
	end := now.Add(time.Hour)
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: contestID, Name: "Started",
		StartTime: start, EndTime: end, Duration: 120, IsActive: true,
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}

	// End-time-only change after the start is rejected, like a start change.
	_, err = svc.UpdateContest(ctx, &domain.UpdateContestRequest{
		Id: contestID, Name: "Started",
		StartTime: start, EndTime: now.Add(2 * time.Hour), Duration: 120, IsActive: true,
	})
	if !errors.Is(err, domain.ErrContestTimeLocked) {
		t.Errorf("UpdateContest() end-change error = %v, want errors.Is ErrContestTimeLocked", err)
	}

	contestRepo.updateErr = errors.New("db down")
	_, err = svc.UpdateContest(ctx, &domain.UpdateContestRequest{
		Id: contestID, Name: "Renamed",
		StartTime: start, EndTime: end, Duration: 120, IsActive: true,
	})
	if err == nil || !strings.Contains(err.Error(), "failed to update contest") {
		t.Errorf("UpdateContest() update error = %v, want wrapped error", err)
	}
}

func TestContestServiceStartContestErrors(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	if err := services.NewContestService(
		&fakeContestRepository{},
		newFakeUserRepository(),
		nil,
	).StartContest(ctx, uuid.New().String()); err == nil || !strings.Contains(err.Error(), "contest not found") {
		t.Errorf("StartContest() on empty store error = %v, want not-found error", err)
	}

	svc, contestRepo, _, problemsRepo := newContestTestService()
	contestID := uuid.New().String()
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: contestID, Name: "No problems repo",
		StartTime: now.Add(time.Hour), EndTime: now.Add(3 * time.Hour), Duration: 120,
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}
	nilProblemsSvc := services.NewContestService(contestRepo, newFakeUserRepository(), nil)
	if err := nilProblemsSvc.StartContest(ctx, contestID); !errors.Is(err, domain.ErrContestProblemsNotConfigured) {
		t.Errorf("StartContest() nil problems repo error = %v, want errors.Is ErrContestProblemsNotConfigured", err)
	}

	problemsRepo.err = errors.New("db down")
	if err := svc.StartContest(ctx, contestID); err == nil || !strings.Contains(err.Error(), "failed to get contest problems") {
		t.Errorf("StartContest() problems error = %v, want wrapped error", err)
	}
	problemsRepo.err = nil

	contestRepo.updateErr = errors.New("db down")
	if err := problemsRepo.CreateContestProblem(ctx, &domain.ContestProblems{
		UniqueID: uuid.New().String(), ContestID: contestID, ProblemID: uuid.New().String(), OrderPosition: 1, MaxPoints: 100,
	}); err != nil {
		t.Fatalf("CreateContestProblem() error = %v", err)
	}
	if err := svc.StartContest(ctx, contestID); err == nil || !strings.Contains(err.Error(), "failed to update contest") {
		t.Errorf("StartContest() update error = %v, want wrapped error", err)
	}
}

func TestContestServiceStartContestRestartActiveIsIdempotent(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, problemsRepo := newContestTestService()
	now := time.Now()
	contestID := uuid.New().String()
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: contestID, Name: "Active",
		StartTime: now.Add(-time.Hour), EndTime: now.Add(time.Hour), Duration: 120, IsActive: true,
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}
	if err := problemsRepo.CreateContestProblem(ctx, &domain.ContestProblems{
		UniqueID: uuid.New().String(), ContestID: contestID, ProblemID: uuid.New().String(), OrderPosition: 1, MaxPoints: 100,
	}); err != nil {
		t.Fatalf("CreateContestProblem() error = %v", err)
	}

	// Pinned behaviour: re-starting an already-active contest succeeds
	// and leaves it active; there is no re-start guard.
	if err := svc.StartContest(ctx, contestID); err != nil {
		t.Fatalf("StartContest() re-start error = %v, want success (idempotent)", err)
	}
	stored, err := contestRepo.GetByID(ctx, contestID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if !stored.IsActive {
		t.Error("StartContest() re-start left IsActive=false")
	}
}

func TestContestServiceEndContestErrorsAndPinnedBehaviour(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()
	now := time.Now()

	if err := svc.EndContest(ctx, uuid.New().String()); !errors.Is(err, domain.ErrContestNotFound) {
		t.Errorf("EndContest() missing error = %v, want errors.Is ErrContestNotFound", err)
	}

	contestID := uuid.New().String()
	end := now.Add(time.Hour)
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: contestID, Name: "Live",
		StartTime: now.Add(-time.Hour), EndTime: end, Duration: 120, IsActive: true,
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}
	contestRepo.updateErr = errors.New("db down")
	if err := svc.EndContest(ctx, contestID); err == nil || !strings.Contains(err.Error(), "failed to update contest") {
		t.Errorf("EndContest() update error = %v, want wrapped error", err)
	}
	contestRepo.updateErr = nil

	// Pinned behaviour: EndContest is unconditional — ending an
	// already-inactive contest, or one past its end window, succeeds
	// and leaves EndTime unchanged.
	inactiveID := uuid.New().String()
	pastID := uuid.New().String()
	pastEnd := now.Add(-time.Hour)
	for _, c := range []domain.Contest{
		{Id: inactiveID, Name: "Inactive", StartTime: now.Add(-2 * time.Hour), EndTime: end, Duration: 120},
		{Id: pastID, Name: "Past", StartTime: now.Add(-3 * time.Hour), EndTime: pastEnd, Duration: 120, IsActive: true},
	} {
		if err := contestRepo.CreateContest(ctx, &c); err != nil {
			t.Fatalf("CreateContest() error = %v", err)
		}
	}
	if err := svc.EndContest(ctx, inactiveID); err != nil {
		t.Errorf("EndContest() inactive error = %v, want success (unconditional)", err)
	}
	if err := svc.EndContest(ctx, pastID); err != nil {
		t.Errorf("EndContest() past-window error = %v, want success (unconditional)", err)
	}
	stored, err := contestRepo.GetByID(ctx, pastID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}
	if stored.IsActive || !stored.EndTime.Equal(pastEnd) {
		t.Errorf("EndContest() past = IsActive %v EndTime %v, want false and %v", stored.IsActive, stored.EndTime, pastEnd)
	}
}

func TestContestServiceDeleteContestErrors(t *testing.T) {
	ctx := context.Background()
	svc, contestRepo, _, _ := newContestTestService()
	now := time.Now()

	if err := svc.DeleteContest(ctx, &domain.DeleteContestRequest{Id: uuid.New().String()}); !errors.Is(err, domain.ErrContestNotFound) {
		t.Errorf("DeleteContest() missing error = %v, want errors.Is ErrContestNotFound", err)
	}

	contestID := uuid.New().String()
	if err := contestRepo.CreateContest(ctx, &domain.Contest{
		Id: contestID, Name: "Deletable",
		StartTime: now.Add(2 * time.Hour), EndTime: now.Add(4 * time.Hour), Duration: 120,
	}); err != nil {
		t.Fatalf("CreateContest() error = %v", err)
	}
	contestRepo.deleteErr = errors.New("db down")
	if err := svc.DeleteContest(ctx, &domain.DeleteContestRequest{Id: contestID}); err == nil || !strings.Contains(err.Error(), "failed to delete contest") {
		t.Errorf("DeleteContest() delete error = %v, want wrapped error", err)
	}
}
