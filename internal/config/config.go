// Package config manages all configuration and credentials for butea-cli.
//
// File layout:
//
//	~/.butea/
//	    config.toml   – global settings (API URL, app URL)
//	    cred.toml     – credentials (access + refresh tokens)  [mode 0600]
//
//	<repo>/
//	    .butea.toml   – per-project settings (project_id, branch)
package config

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// ── Defaults ─────────────────────────────────────────────────────────────────

const (
	DefaultAPIURL = "https://api.butea.app"
	DefaultAppURL = "https://app.butea.app"

	// EnvAPIURL and EnvAppURL are the environment variable names that override
	// the stored config. Useful in CI, Docker, or self-hosted deployments.
	EnvAPIURL = "BUTEA_API_URL"
	EnvAppURL = "BUTEA_APP_URL"
)

// ErrNotLoggedIn is returned when an auth-required operation is attempted
// without stored credentials.
var ErrNotLoggedIn = errors.New("not logged in – run 'butea init' to authenticate")

// ── Directory helpers ─────────────────────────────────────────────────────────

// GlobalDir returns the ~/.butea directory path.
func GlobalDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".butea"), nil
}

// GlobalConfigPath returns the path to ~/.butea/config.toml.
func GlobalConfigPath() (string, error) {
	d, err := GlobalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "config.toml"), nil
}

// CredPath returns the path to ~/.butea/cred.toml.
func CredPath() (string, error) {
	d, err := GlobalDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, "cred.toml"), nil
}

// LocalConfigPath returns .butea.toml in the current working directory.
func LocalConfigPath() string {
	return ".butea.toml"
}

// ── Global config ─────────────────────────────────────────────────────────────

// GlobalConfig holds platform-wide settings (~/.butea/config.toml).
type GlobalConfig struct {
	APIURL string `toml:"api_url"`
	AppURL string `toml:"app_url"`
}

// LoadGlobal reads ~/.butea/config.toml, then overlays BUTEA_API_URL /
// BUTEA_APP_URL environment variables (if set).
// Returns defaults when the file does not yet exist.
func LoadGlobal() (*GlobalConfig, error) {
	path, err := GlobalConfigPath()
	if err != nil {
		return nil, err
	}
	cfg := &GlobalConfig{APIURL: DefaultAPIURL, AppURL: DefaultAppURL}
	data, err := os.ReadFile(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	if err == nil {
		if _, decErr := toml.Decode(string(data), cfg); decErr != nil {
			return nil, decErr
		}
	}
	// Backfill empty values from defaults
	if cfg.APIURL == "" {
		cfg.APIURL = DefaultAPIURL
	}
	if cfg.AppURL == "" {
		cfg.AppURL = DefaultAppURL
	}
	// Env vars always win over the file
	if v := os.Getenv(EnvAPIURL); v != "" {
		cfg.APIURL = v
	}
	if v := os.Getenv(EnvAppURL); v != "" {
		cfg.AppURL = v
	}
	return cfg, nil
}

// Save writes the global config to ~/.butea/config.toml (mode 0644).
func (c *GlobalConfig) Save() error {
	dir, err := GlobalDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, "config.toml")
	return writeTOML(path, c, 0644)
}

// ── Credentials ───────────────────────────────────────────────────────────────

// Credentials holds auth tokens (~/.butea/cred.toml).
type Credentials struct {
	AccessToken  string `toml:"access_token"`
	RefreshToken string `toml:"refresh_token"`
}

// LoadCredentials reads ~/.butea/cred.toml.
// Returns an empty Credentials (no tokens) when the file does not exist.
func LoadCredentials() (*Credentials, error) {
	path, err := CredPath()
	if err != nil {
		return nil, err
	}
	cred := &Credentials{}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cred, nil
	}
	if err != nil {
		return nil, err
	}
	if _, err := toml.Decode(string(data), cred); err != nil {
		return nil, err
	}
	return cred, nil
}

// Save writes credentials to ~/.butea/cred.toml (mode 0600, owner-only).
func (c *Credentials) Save() error {
	dir, err := GlobalDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	path := filepath.Join(dir, "cred.toml")
	return writeTOML(path, c, 0600)
}

// Clear zeroes tokens and saves.
func (c *Credentials) Clear() error {
	c.AccessToken = ""
	c.RefreshToken = ""
	return c.Save()
}

// IsLoggedIn reports whether credentials hold a non-empty access token.
func (c *Credentials) IsLoggedIn() bool {
	return c.AccessToken != ""
}

// RequireAuth returns ErrNotLoggedIn when no access token is stored.
func (c *Credentials) RequireAuth() error {
	if !c.IsLoggedIn() {
		return ErrNotLoggedIn
	}
	return nil
}

// ── Local project config ──────────────────────────────────────────────────────

// LocalConfig is the per-repository config stored in .butea.toml.
type LocalConfig struct {
	ProjectID string `toml:"project_id"`
	Branch    string `toml:"branch,omitempty"`
}

// LoadLocal reads .butea.toml from the current directory.
// Returns nil, nil when the file does not exist.
func LoadLocal() (*LocalConfig, error) {
	data, err := os.ReadFile(LocalConfigPath())
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var lc LocalConfig
	if _, err := toml.Decode(string(data), &lc); err != nil {
		return nil, err
	}
	return &lc, nil
}

// Save writes .butea.toml to the current directory (mode 0644).
func (l *LocalConfig) Save() error {
	return writeTOML(LocalConfigPath(), l, 0644)
}

// ── TOML write helper ─────────────────────────────────────────────────────────

func writeTOML(path string, v any, mode os.FileMode) error {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if err := toml.NewEncoder(f).Encode(v); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
