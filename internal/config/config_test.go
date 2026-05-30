package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ButeaLabs/butea-cli/internal/config"
)

// withTempHome overrides $HOME so all config files land in a temp dir.
func withTempHome(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	return tmp
}

// ── GlobalConfig ─────────────────────────────────────────────────────────────

func TestLoadGlobal_DefaultsWhenMissing(t *testing.T) {
	withTempHome(t)
	cfg, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if cfg.APIURL != config.DefaultAPIURL {
		t.Errorf("APIURL = %q; want %q", cfg.APIURL, config.DefaultAPIURL)
	}
	if cfg.AppURL != config.DefaultAppURL {
		t.Errorf("AppURL = %q; want %q", cfg.AppURL, config.DefaultAppURL)
	}
}

func TestGlobalConfig_SaveAndLoad(t *testing.T) {
	withTempHome(t)
	orig := &config.GlobalConfig{
		APIURL: "https://api.mycompany.com",
		AppURL: "https://app.mycompany.com",
	}
	if err := orig.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if got.APIURL != orig.APIURL {
		t.Errorf("APIURL = %q; want %q", got.APIURL, orig.APIURL)
	}
	if got.AppURL != orig.AppURL {
		t.Errorf("AppURL = %q; want %q", got.AppURL, orig.AppURL)
	}
}

func TestGlobalConfig_EnvVarsOverrideFile(t *testing.T) {
	withTempHome(t)
	// Write a config file with one set of URLs
	file := &config.GlobalConfig{
		APIURL: "https://api.from-file.com",
		AppURL: "https://app.from-file.com",
	}
	if err := file.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	// Set env vars to different values
	t.Setenv(config.EnvAPIURL, "https://api.from-env.com")
	t.Setenv(config.EnvAppURL, "https://app.from-env.com")

	cfg, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if cfg.APIURL != "https://api.from-env.com" {
		t.Errorf("APIURL = %q; want env value", cfg.APIURL)
	}
	if cfg.AppURL != "https://app.from-env.com" {
		t.Errorf("AppURL = %q; want env value", cfg.AppURL)
	}
}

func TestGlobalConfig_EnvVarsPartialOverride(t *testing.T) {
	withTempHome(t)
	file := &config.GlobalConfig{
		APIURL: "https://api.from-file.com",
		AppURL: "https://app.from-file.com",
	}
	if err := file.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	// Only override the API URL
	t.Setenv(config.EnvAPIURL, "https://api.override.com")

	cfg, err := config.LoadGlobal()
	if err != nil {
		t.Fatalf("LoadGlobal() error = %v", err)
	}
	if cfg.APIURL != "https://api.override.com" {
		t.Errorf("APIURL = %q; want env override", cfg.APIURL)
	}
	// AppURL comes from file, not overridden
	if cfg.AppURL != "https://app.from-file.com" {
		t.Errorf("AppURL = %q; want file value", cfg.AppURL)
	}
}

// ── Credentials ───────────────────────────────────────────────────────────────

func TestLoadCredentials_EmptyWhenMissing(t *testing.T) {
	withTempHome(t)
	cred, err := config.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}
	if cred.AccessToken != "" || cred.RefreshToken != "" {
		t.Error("credentials should be empty on first load")
	}
}

func TestCredentials_SaveAndLoad(t *testing.T) {
	withTempHome(t)
	orig := &config.Credentials{AccessToken: "acc_tok", RefreshToken: "ref_tok"}
	if err := orig.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := config.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}
	if got.AccessToken != orig.AccessToken {
		t.Errorf("AccessToken = %q; want %q", got.AccessToken, orig.AccessToken)
	}
	if got.RefreshToken != orig.RefreshToken {
		t.Errorf("RefreshToken = %q; want %q", got.RefreshToken, orig.RefreshToken)
	}
}

func TestCredentials_FilePermissions(t *testing.T) {
	withTempHome(t)
	cred := &config.Credentials{AccessToken: "secret", RefreshToken: "also-secret"}
	if err := cred.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	path, err := config.CredPath()
	if err != nil {
		t.Fatalf("CredPath() error = %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("cred.toml permissions = %04o; want 0600", info.Mode().Perm())
	}
}

func TestCredentials_Clear(t *testing.T) {
	withTempHome(t)
	cred := &config.Credentials{AccessToken: "acc", RefreshToken: "ref"}
	if err := cred.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if err := cred.Clear(); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}
	got, err := config.LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials() error = %v", err)
	}
	if got.AccessToken != "" {
		t.Errorf("AccessToken after Clear = %q; want empty", got.AccessToken)
	}
}

func TestCredentials_RequireAuth(t *testing.T) {
	t.Run("no token", func(t *testing.T) {
		c := &config.Credentials{}
		if err := c.RequireAuth(); err == nil {
			t.Error("RequireAuth() should error with no token")
		}
	})
	t.Run("with token", func(t *testing.T) {
		c := &config.Credentials{AccessToken: "tok"}
		if err := c.RequireAuth(); err != nil {
			t.Errorf("RequireAuth() error = %v; want nil", err)
		}
	})
}

func TestCredentials_IsLoggedIn(t *testing.T) {
	if (&config.Credentials{}).IsLoggedIn() {
		t.Error("IsLoggedIn() should be false with empty token")
	}
	if !(&config.Credentials{AccessToken: "x"}).IsLoggedIn() {
		t.Error("IsLoggedIn() should be true with a token")
	}
}

// ── GlobalDir / CredPath structure ────────────────────────────────────────────

func TestGlobalDir_IsUnderHome(t *testing.T) {
	tmp := withTempHome(t)
	dir, err := config.GlobalDir()
	if err != nil {
		t.Fatalf("GlobalDir() error = %v", err)
	}
	expected := filepath.Join(tmp, ".butea")
	if dir != expected {
		t.Errorf("GlobalDir() = %q; want %q", dir, expected)
	}
}

func TestGlobalConfig_CreatesDirectory(t *testing.T) {
	withTempHome(t)
	cfg := &config.GlobalConfig{APIURL: config.DefaultAPIURL}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	path, _ := config.GlobalConfigPath()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("config.toml not created: %v", err)
	}
}

// ── LocalConfig ───────────────────────────────────────────────────────────────

func TestLoadLocal_NilWhenMissing(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })
	os.Chdir(tmp)

	lc, err := config.LoadLocal()
	if err != nil {
		t.Fatalf("LoadLocal() error = %v; want nil", err)
	}
	if lc != nil {
		t.Error("LoadLocal() should return nil when .butea.toml is missing")
	}
}

func TestLocalConfig_SaveAndLoad(t *testing.T) {
	tmp := t.TempDir()
	orig, _ := os.Getwd()
	t.Cleanup(func() { os.Chdir(orig) })
	os.Chdir(tmp)

	orig2 := &config.LocalConfig{ProjectID: "proj-uuid", Branch: "main"}
	if err := orig2.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	got, err := config.LoadLocal()
	if err != nil {
		t.Fatalf("LoadLocal() error = %v", err)
	}
	if got == nil {
		t.Fatal("LoadLocal() returned nil after Save()")
	}
	if got.ProjectID != orig2.ProjectID {
		t.Errorf("ProjectID = %q; want %q", got.ProjectID, orig2.ProjectID)
	}
	if got.Branch != orig2.Branch {
		t.Errorf("Branch = %q; want %q", got.Branch, orig2.Branch)
	}
}
