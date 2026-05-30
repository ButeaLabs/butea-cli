package api_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ButeaLabs/butea-cli/internal/api"
)

// newTestServer spins up an httptest.Server that responds to a single path+method
// with the given status code and JSON body.
func newTestServer(t *testing.T, method, path string, status int, body any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method || r.URL.Path != path {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if body != nil {
			_ = json.NewEncoder(w).Encode(body)
		}
	}))
}

// tc creates a test Client with a fixed "test" version string.
func tc(baseURL, access, refresh string) *api.Client {
	return api.NewClient(baseURL, access, refresh, "test")
}

// ── Login ────────────────────────────────────────────────────────────────────

func TestLogin_Success(t *testing.T) {
	want := api.TokenPairResponse{
		AccessToken:  "acc_123",
		RefreshToken: "ref_456",
		TokenType:    "Bearer",
		ExpiresIn:    900,
	}
	srv := newTestServer(t, http.MethodPost, "/auth/login", http.StatusOK, want)
	defer srv.Close()

	client := tc(srv.URL, "", "")
	got, err := client.Login(context.Background(), "user@example.com", "pass")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if got.AccessToken != want.AccessToken {
		t.Errorf("AccessToken = %q; want %q", got.AccessToken, want.AccessToken)
	}
	if got.RefreshToken != want.RefreshToken {
		t.Errorf("RefreshToken = %q; want %q", got.RefreshToken, want.RefreshToken)
	}
}

func TestLogin_BadCredentials(t *testing.T) {
	srv := newTestServer(t, http.MethodPost, "/auth/login", http.StatusUnauthorized,
		api.APIError{Code: "UNAUTHORIZED", Message: "invalid credentials"})
	defer srv.Close()

	client := tc(srv.URL, "", "")
	_, err := client.Login(context.Background(), "bad@example.com", "wrong")
	if err == nil {
		t.Fatal("Login() should return error on 401")
	}
}

// ── GetMe ────────────────────────────────────────────────────────────────────

func TestGetMe_Success(t *testing.T) {
	name := "Test User"
	want := api.UserProfile{
		ID:       "uuid-1",
		Email:    "user@example.com",
		Name:     &name,
		IsActive: true,
	}
	srv := newTestServer(t, http.MethodGet, "/api/v1/me", http.StatusOK, want)
	defer srv.Close()

	client := tc(srv.URL, "acc_tok", "ref_tok")
	got, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe() error = %v", err)
	}
	if got.Email != want.Email {
		t.Errorf("Email = %q; want %q", got.Email, want.Email)
	}
}

func TestGetMe_Unauthorized(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/auth/refresh" {
			// Refresh fails too
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(api.APIError{Code: "UNAUTHORIZED", Message: "unauthorized"})
	}))
	defer srv.Close()

	// Client has a refresh token but the refresh endpoint also fails
	client := tc(srv.URL, "expired_token", "bad_refresh")
	_, err := client.GetMe(context.Background())
	if err == nil {
		t.Fatal("GetMe() should return error on persistent 401")
	}
}

// ── Token refresh ────────────────────────────────────────────────────────────

func TestAutoRefresh_RetriesAfter401(t *testing.T) {
	var refreshCalled bool
	var retryCount int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/me":
			retryCount++
			if retryCount == 1 {
				// First call: simulate expired token
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			// Second call (after refresh): success
			name := "Refreshed User"
			_ = json.NewEncoder(w).Encode(api.UserProfile{ID: "u1", Email: "e@e.com", Name: &name})
		case "/auth/refresh":
			refreshCalled = true
			_ = json.NewEncoder(w).Encode(api.TokenPairResponse{
				AccessToken:  "new_acc",
				RefreshToken: "new_ref",
			})
		}
	}))
	defer srv.Close()

	var savedAccess string
	client := tc(srv.URL, "old_acc", "old_ref")
	client.OnTokenRefresh = func(a, _ string) { savedAccess = a }

	_, err := client.GetMe(context.Background())
	if err != nil {
		t.Fatalf("GetMe() error = %v; expected auto-refresh to succeed", err)
	}
	if !refreshCalled {
		t.Error("refresh endpoint was not called")
	}
	if savedAccess != "new_acc" {
		t.Errorf("OnTokenRefresh received %q; want 'new_acc'", savedAccess)
	}
}

// ── ListProjects ─────────────────────────────────────────────────────────────

func TestListProjects_Success(t *testing.T) {
	want := api.ProjectListResponse{
		Projects: []api.Project{
			{ID: "p1", Name: "Alpha", Provider: "github", DefaultBranch: "main"},
			{ID: "p2", Name: "Beta", Provider: "gitlab", DefaultBranch: "develop"},
		},
		Total: 2,
	}
	srv := newTestServer(t, http.MethodGet, "/api/v1/projects/", http.StatusOK, want)
	defer srv.Close()

	client := tc(srv.URL, "tok", "")
	got, err := client.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(got.Projects) != 2 {
		t.Errorf("len(Projects) = %d; want 2", len(got.Projects))
	}
	if got.Projects[0].Name != "Alpha" {
		t.Errorf("Projects[0].Name = %q; want 'Alpha'", got.Projects[0].Name)
	}
}

// ── CreateDeployment ─────────────────────────────────────────────────────────

func TestCreateDeployment_Success(t *testing.T) {
	pid := "proj-uuid"
	want := api.DeploymentResponse{
		Deployment: &api.Deployment{
			ID:        "dep-1",
			ProjectID: pid,
			Branch:    "main",
			Status:    "queued",
			CreatedAt: time.Now(),
		},
	}
	srv := newTestServer(t, http.MethodPost, "/api/v1/projects/"+pid+"/deployments",
		http.StatusCreated, want)
	defer srv.Close()

	client := tc(srv.URL, "tok", "")
	got, err := client.CreateDeployment(context.Background(), pid, "main")
	if err != nil {
		t.Fatalf("CreateDeployment() error = %v", err)
	}
	if got.Branch != "main" {
		t.Errorf("Branch = %q; want 'main'", got.Branch)
	}
	if got.Status != "queued" {
		t.Errorf("Status = %q; want 'queued'", got.Status)
	}
}

// ── Health ───────────────────────────────────────────────────────────────────

func TestHealth_Healthy(t *testing.T) {
	want := api.HealthResponse{
		Status:  "healthy",
		Version: "1.0.0",
		Services: map[string]api.ServiceHealth{
			"postgreSQL": {Status: "healthy", Latency: "1ms"},
			"redis":      {Status: "healthy", Latency: "0ms"},
		},
	}
	srv := newTestServer(t, http.MethodGet, "/health", http.StatusOK, want)
	defer srv.Close()

	client := tc(srv.URL, "", "")
	got, err := client.Health(context.Background())
	if err != nil {
		t.Fatalf("Health() error = %v", err)
	}
	if got.Status != "healthy" {
		t.Errorf("Status = %q; want 'healthy'", got.Status)
	}
	if _, ok := got.Services["postgreSQL"]; !ok {
		t.Error("Services['postgreSQL'] missing")
	}
}

func TestHealth_Degraded(t *testing.T) {
	want := api.HealthResponse{
		Status: "degraded",
		Services: map[string]api.ServiceHealth{
			"postgreSQL": {Status: "unhealthy", Error: "connection refused"},
		},
	}
	srv := newTestServer(t, http.MethodGet, "/health", http.StatusServiceUnavailable, want)
	defer srv.Close()

	// The BE returns 503 for degraded health. The client treats any >=400 as an
	// APIError — the health *command* handles this gracefully, but the client
	// itself must surface the error so callers can detect it.
	client := tc(srv.URL, "", "")
	_, err := client.Health(context.Background())
	if err == nil {
		t.Error("Health() should return an error when the server responds with 503")
	}
}

// ── APIError ──────────────────────────────────────────────────────────────────

func TestAPIError_Format(t *testing.T) {
	e := &api.APIError{Code: "NOT_FOUND", Message: "resource not found"}
	if e.Error() != "NOT_FOUND: resource not found" {
		t.Errorf("Error() = %q; unexpected format", e.Error())
	}
}

func TestAPIError_NoCode(t *testing.T) {
	e := &api.APIError{Message: "something went wrong"}
	if e.Error() != "something went wrong" {
		t.Errorf("Error() = %q; expected plain message", e.Error())
	}
}

// ── CancelDeployment ─────────────────────────────────────────────────────────

func TestCancelDeployment_Success(t *testing.T) {
	depID := "dep-uuid"
	srv := newTestServer(t, http.MethodPost, "/api/v1/deployments/"+depID+"/cancel",
		http.StatusNoContent, nil)
	defer srv.Close()

	client := tc(srv.URL, "tok", "")
	if err := client.CancelDeployment(context.Background(), depID); err != nil {
		t.Fatalf("CancelDeployment() error = %v", err)
	}
}

func TestDeleteProject_Success(t *testing.T) {
	projID := "proj-uuid"
	srv := newTestServer(t, http.MethodDelete, "/api/v1/projects/"+projID,
		http.StatusNoContent, nil)
	defer srv.Close()

	client := tc(srv.URL, "tok", "")
	if err := client.DeleteProject(context.Background(), projID); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}
}

