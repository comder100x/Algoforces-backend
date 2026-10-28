package router_test

import (
	"algoforces/internal/domain"
	"algoforces/internal/handlers"
	"algoforces/internal/router"
	"algoforces/internal/utils"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// fakeContestUseCase serves one visible contest for GetVisibleContests;
// every other method is an unused stub for routing purposes.
type fakeContestUseCase struct{}

func (fakeContestUseCase) CreateContest(_ context.Context, _ *domain.CreateContestRequest, _ string) (*domain.CreateContestResponse, error) {
	return nil, errors.New("not used")
}

func (fakeContestUseCase) UpdateContest(_ context.Context, _ *domain.UpdateContestRequest) (*domain.UpdateContestResponse, error) {
	return nil, errors.New("not used")
}

func (fakeContestUseCase) GetContestDetails(_ context.Context, _ string) (*domain.CreateContestResponse, error) {
	return nil, errors.New("not used")
}

func (fakeContestUseCase) DeleteContest(_ context.Context, _ *domain.DeleteContestRequest) error {
	return errors.New("not used")
}

func (fakeContestUseCase) StartContest(_ context.Context, _ string) error {
	return errors.New("not used")
}

func (fakeContestUseCase) EndContest(_ context.Context, _ string) error {
	return errors.New("not used")
}

func (fakeContestUseCase) GetAllContests(_ context.Context) ([]domain.CreateContestResponse, error) {
	return nil, errors.New("not used")
}

func (fakeContestUseCase) GetVisibleContests(_ context.Context) ([]domain.CreateContestResponse, error) {
	return []domain.CreateContestResponse{
		{Id: "visible-1", Name: "Visible", Status: domain.ContestStatusUpcoming, Visible: true},
	}, nil
}

func newTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	// Other handlers are nil: their routes are registered but never
	// invoked by these tests, so only the contest handler is real.
	return router.New(router.Handlers{
		Contest: handlers.NewContestHandler(fakeContestUseCase{}),
	})
}

// TestGetVisibleContestsRequiresAuthButNoRoleGate encodes the corrected
// review outcome for GET /api/contest/all: authentication IS required,
// but after authentication every role may access it — there is no role
// gate on this route. It runs against the real gin engine via httptest;
// no live server or database is involved.
func TestGetVisibleContestsRequiresAuthButNoRoleGate(t *testing.T) {
	engine := newTestEngine()

	// No token: rejected.
	req := httptest.NewRequest(http.MethodGet, "/api/contest/all", nil)
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("GET /api/contest/all without token = %d, want %d (auth required)", w.Code, http.StatusUnauthorized)
	}

	// Any authenticated role: allowed, and the enum status still
	// renders as the plain JSON string.
	for _, role := range []string{"user", "admin", "problem-setter"} {
		token, err := utils.GenerateToken("user-1", role, "user@example.com")
		if err != nil {
			t.Fatalf("GenerateToken(%q) error = %v", role, err)
		}
		req := httptest.NewRequest(http.MethodGet, "/api/contest/all", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		engine.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("GET /api/contest/all as role %q = %d, want %d", role, w.Code, http.StatusOK)
			continue
		}
		var body struct {
			Data []struct {
				ID     string `json:"id"`
				Status string `json:"status"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("GET /api/contest/all as role %q: unmarshal error = %v", role, err)
		}
		if len(body.Data) != 1 || body.Data[0].ID != "visible-1" || body.Data[0].Status != "upcoming" {
			t.Errorf("GET /api/contest/all as role %q: data = %+v, want one contest with status %q", role, body.Data, "upcoming")
		}
	}
}
