// Package api provides a typed HTTP client for the Butea API.
// Every method maps 1-to-1 to a BE endpoint documented in sal/internal/router/router.go.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client is the Butea API HTTP client.
// It automatically retries once after a 401 by refreshing the access token.
type Client struct {
	baseURL      string
	accessToken  string
	refreshToken string
	userAgent    string
	http         *http.Client
	// OnTokenRefresh is called after a successful silent token refresh so the
	// caller can persist the new tokens.
	OnTokenRefresh func(accessToken, refreshToken string)
}

// NewClient creates a new Client. version is included in the User-Agent header
// (e.g. "butea-cli/v1.2.3") and is injected at build time via ldflags.
func NewClient(baseURL, accessToken, refreshToken, version string) *Client {
	return &Client{
		baseURL:      baseURL,
		accessToken:  accessToken,
		refreshToken: refreshToken,
		userAgent:    "butea-cli/" + version,
		http:         &http.Client{Timeout: 30 * time.Second},
	}
}

// ── Auth ─────────────────────────────────────────────────────────────────────

// Login authenticates with email + password and returns a token pair.
// Corresponds to: POST /auth/login
func (c *Client) Login(ctx context.Context, email, password string) (*TokenPairResponse, error) {
	var out TokenPairResponse
	if err := c.do(ctx, http.MethodPost, "/auth/login",
		LoginInput{Email: email, Password: password}, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Logout revokes the given refresh token.
// Corresponds to: POST /auth/logout  (auth-required)
func (c *Client) Logout(ctx context.Context, refreshToken string) error {
	return c.do(ctx, http.MethodPost, "/auth/logout",
		map[string]string{"refresh_token": refreshToken}, nil)
}

// GetMe returns the authenticated user's profile.
// Corresponds to: GET /api/v1/me
func (c *Client) GetMe(ctx context.Context) (*UserProfile, error) {
	var out UserProfile
	if err := c.do(ctx, http.MethodGet, "/api/v1/me", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ── Projects ─────────────────────────────────────────────────────────────────

// ListProjects returns all projects visible to the authenticated user.
// Corresponds to: GET /api/v1/projects/
func (c *Client) ListProjects(ctx context.Context) (*ProjectListResponse, error) {
	var out ProjectListResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/projects/", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetProject returns a single project by ID.
// Corresponds to: GET /api/v1/projects/{projectId}
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	var out ProjectResponse
	if err := c.do(ctx, http.MethodGet, "/api/v1/projects/"+projectID, nil, &out); err != nil {
		return nil, err
	}
	return out.Project, nil
}

// DeleteProject deletes a project.
// Corresponds to: DELETE /api/v1/projects/{projectId}
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	return c.do(ctx, http.MethodDelete, "/api/v1/projects/"+projectID, nil, nil)
}

// ── Deployments ──────────────────────────────────────────────────────────────

// CreateDeployment triggers a new deployment for a project on the given branch.
// Corresponds to: POST /api/v1/projects/{projectId}/deployments
func (c *Client) CreateDeployment(ctx context.Context, projectID, branch string) (*Deployment, error) {
	var out DeploymentResponse
	if err := c.do(ctx, http.MethodPost,
		"/api/v1/projects/"+projectID+"/deployments",
		CreateDeploymentInput{Branch: branch}, &out); err != nil {
		return nil, err
	}
	return out.Deployment, nil
}

// ListDeployments returns paginated deployments for a project.
// Corresponds to: GET /api/v1/projects/{projectId}/deployments
func (c *Client) ListDeployments(ctx context.Context, projectID string) (*DeploymentListResponse, error) {
	var out DeploymentListResponse
	if err := c.do(ctx, http.MethodGet,
		"/api/v1/projects/"+projectID+"/deployments", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetDeployment returns a single deployment by ID.
// Corresponds to: GET /api/v1/deployments/{deploymentId}
func (c *Client) GetDeployment(ctx context.Context, deploymentID string) (*Deployment, error) {
	var out DeploymentResponse
	if err := c.do(ctx, http.MethodGet,
		"/api/v1/deployments/"+deploymentID, nil, &out); err != nil {
		return nil, err
	}
	return out.Deployment, nil
}

// CancelDeployment cancels a pending/queued deployment.
// Corresponds to: POST /api/v1/deployments/{deploymentId}/cancel
func (c *Client) CancelDeployment(ctx context.Context, deploymentID string) error {
	return c.do(ctx, http.MethodPost,
		"/api/v1/deployments/"+deploymentID+"/cancel", nil, nil)
}

// ── Health ───────────────────────────────────────────────────────────────────

// Health pings the /health endpoint and returns the server status.
func (c *Client) Health(ctx context.Context) (*HealthResponse, error) {
	var out HealthResponse
	if err := c.do(ctx, http.MethodGet, "/health", nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ── low-level HTTP ───────────────────────────────────────────────────────────

func (c *Client) do(ctx context.Context, method, path string, body, out any) error {
	return c.doOnce(ctx, method, path, body, out, true)
}

func (c *Client) doOnce(ctx context.Context, method, path string, body, out any, allowRefresh bool) error {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Auto-refresh on 401 if we have a refresh token and haven't already tried.
	if resp.StatusCode == http.StatusUnauthorized && allowRefresh && c.refreshToken != "" {
		newAccess, newRefresh, refreshErr := c.refresh(ctx)
		if refreshErr == nil {
			c.accessToken = newAccess
			c.refreshToken = newRefresh
			if c.OnTokenRefresh != nil {
				c.OnTokenRefresh(newAccess, newRefresh)
			}
			return c.doOnce(ctx, method, path, body, out, false)
		}
	}

	if resp.StatusCode >= 400 {
		var apiErr APIError
		_ = json.NewDecoder(resp.Body).Decode(&apiErr)
		if apiErr.Message == "" {
			apiErr.Message = http.StatusText(resp.StatusCode)
			apiErr.Code = fmt.Sprintf("HTTP_%d", resp.StatusCode)
		}
		return &apiErr
	}

	if out != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// refresh silently rotates the access token using the stored refresh token.
func (c *Client) refresh(ctx context.Context) (string, string, error) {
	payload, _ := json.Marshal(map[string]string{"refresh_token": c.refreshToken})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/auth/refresh", bytes.NewReader(payload))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("token refresh returned HTTP %d", resp.StatusCode)
	}
	var tokens TokenPairResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return "", "", err
	}
	return tokens.AccessToken, tokens.RefreshToken, nil
}

