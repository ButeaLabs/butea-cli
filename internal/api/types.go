package api

import "time"

// ── Auth ─────────────────────────────────────────────────────────────────────

// LoginInput is the request body for POST /auth/login.
type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// TokenPairResponse mirrors sal/internal/dto.TokenPairResponse.
type TokenPairResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// ── Users ─────────────────────────────────────────────────────────────────────

// UserProfile mirrors sal/internal/dto.UserProfileResponse.
type UserProfile struct {
	ID        string          `json:"id"`
	Email     string          `json:"email"`
	Name      *string         `json:"name"`
	AvatarURL *string         `json:"avatar_url"`
	IsActive  bool            `json:"is_active"`
	Providers []OAuthProvider `json:"providers"`
	CreatedAt time.Time       `json:"created_at"`
}

// OAuthProvider mirrors sal/internal/dto.OAuthProvider.
type OAuthProvider struct {
	Provider    string `json:"provider"`
	Email       string `json:"email,omitempty"`
	Name        string `json:"name,omitempty"`
	InstanceURL string `json:"instance_url,omitempty"`
}

// ── Projects ─────────────────────────────────────────────────────────────────

// Project mirrors sal/internal/dto.Project.
type Project struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Slug          string     `json:"slug"`
	Description   *string    `json:"description"`
	Provider      string     `json:"provider"`
	RepoFullName  string     `json:"repo_full_name"`
	DefaultBranch string     `json:"default_branch"`
	IsPrivate     bool       `json:"is_private"`
	Status        string     `json:"status"`
	Source        string     `json:"source"`
	Framework     *string    `json:"framework"`
	CloneURL      string     `json:"clone_url"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// ProjectResponse wraps a single project (matches BE envelope).
type ProjectResponse struct {
	Project *Project `json:"project"`
}

// ProjectListResponse mirrors sal/internal/dto.ProjectListResponse.
type ProjectListResponse struct {
	Projects []Project `json:"projects"`
	Total    int       `json:"total"`
	Page     int       `json:"page"`
	PerPage  int       `json:"per_page"`
}

// ── Deployments ──────────────────────────────────────────────────────────────

// CreateDeploymentInput mirrors sal/internal/dto.CreateDeploymentInput
// (only the fields the CLI needs to send).
type CreateDeploymentInput struct {
	Branch string `json:"branch"`
}

// Deployment mirrors sal/internal/dto.Deployment.
type Deployment struct {
	ID             string    `json:"id"`
	ProjectID      string    `json:"project_id"`
	Branch         string    `json:"branch"`
	CommitSHA      string    `json:"commit_sha"`
	CommitMessage  string    `json:"commit_message"`
	Status         string    `json:"status"`
	DeployURL      string    `json:"deploy_url"`
	ErrorMessage   *string   `json:"error_message,omitempty"`
	Framework      string    `json:"framework"`
	Source         string    `json:"source"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// DeploymentResponse wraps a single deployment.
type DeploymentResponse struct {
	Deployment *Deployment `json:"deployment"`
}

// DeploymentListResponse mirrors sal/internal/dto.DeploymentListResponse.
type DeploymentListResponse struct {
	Deployments []Deployment `json:"deployments"`
	Total       int          `json:"total"`
	Page        int          `json:"page"`
	PerPage     int          `json:"per_page"`
}

// ── Health ───────────────────────────────────────────────────────────────────

// ServiceHealth mirrors sal/internal/handler.ServiceHealth.
type ServiceHealth struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
	Latency string `json:"latency,omitempty"`
	Error   string `json:"error,omitempty"`
}

// HealthResponse mirrors sal/internal/handler.HealthResponse.
type HealthResponse struct {
	Status    string                   `json:"status"`
	Timestamp time.Time                `json:"timestamp"`
	Version   string                   `json:"version"`
	GoVersion string                   `json:"goVersion"`
	Services  map[string]ServiceHealth `json:"services"`
}

// ── Errors ───────────────────────────────────────────────────────────────────

// APIError mirrors sal/internal/dto.ErrorResponse.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Code != "" {
		return e.Code + ": " + e.Message
	}
	return e.Message
}

