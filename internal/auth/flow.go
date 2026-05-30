// Package auth handles the browser-based OAuth authentication flow for the CLI.
//
// Flow:
//  1. CLI generates a random state token for CSRF protection.
//  2. CLI starts a local HTTP server on 127.0.0.1 (random free port).
//  3. CLI opens the browser to:
//     https://app.butea.app/signin?cli_port=PORT&cli_state=STATE
//  4. User signs in via the web app (OAuth or email/password).
//  5. The web app's /auth/callback page detects cli_port/cli_state stored in
//     sessionStorage and redirects to:
//     http://127.0.0.1:PORT/auth/callback?access_token=AT&refresh_token=RT&state=STATE
//  6. CLI validates the state, returns tokens, and shuts down the local server.
package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

// Result is returned from StartBrowserFlow after successful authentication.
type Result struct {
	AccessToken  string
	RefreshToken string
}

// StartBrowserFlow runs the browser-based auth handshake.
// It blocks until the user completes authentication or the context is cancelled.
// timeout is the maximum time to wait for the browser callback (0 = 5 minutes).
func StartBrowserFlow(ctx context.Context, appURL string, timeout time.Duration) (*Result, error) {
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	// 1. Generate CSRF state token.
	state, err := generateState()
	if err != nil {
		return nil, fmt.Errorf("generate state: %w", err)
	}

	// 2. Start local listener on a random free loopback port.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("start local server: %w", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	resultCh := make(chan *Result, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Handler: mux, ReadTimeout: 10 * time.Second}

	mux.HandleFunc("/auth/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if q.Get("state") != state {
			http.Error(w, "invalid state token", http.StatusBadRequest)
			errCh <- fmt.Errorf("invalid state token in callback — possible CSRF")
			return
		}

		if errMsg := q.Get("error"); errMsg != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, callbackPageHTML("✗ Authentication failed", errMsg, false))
			errCh <- fmt.Errorf("authentication failed: %s", errMsg)
			go shutdownSoon(srv)
			return
		}

		at := q.Get("access_token")
		rt := q.Get("refresh_token")
		if at == "" || rt == "" {
			http.Error(w, "missing tokens", http.StatusBadRequest)
			errCh <- fmt.Errorf("callback missing access_token or refresh_token")
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, callbackPageHTML("✓ Authenticated", "You can close this tab and return to your terminal.", true))

		resultCh <- &Result{AccessToken: at, RefreshToken: rt}
		go shutdownSoon(srv)
	})

	// Serve in background.
	go func() { _ = srv.Serve(ln) }()

	// 3. Build and print the auth URL.
	authURL := fmt.Sprintf("%s/signin?cli_port=%d&cli_state=%s", appURL, port, state)
	fmt.Printf("  Opening browser for authentication…\n")
	fmt.Printf("  If the browser doesn't open, visit:\n\n    %s\n\n", authURL)

	// 4. Open the browser.
	_ = openBrowser(authURL)

	// 5. Wait for callback.
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case res := <-resultCh:
		return res, nil
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		srv.Shutdown(context.Background()) //nolint:errcheck
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("authentication timed out after %s", timeout)
		}
		return nil, ctx.Err()
	}
}

// openBrowser opens the given URL in the system default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

// generateState creates a cryptographically random hex state string.
func generateState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// shutdownSoon gives the HTTP handler time to finish writing its response
// before shutting down the server.
func shutdownSoon(srv *http.Server) {
	time.Sleep(200 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

// callbackPageHTML returns a minimal self-contained HTML page shown in the
// browser after the auth callback is processed.
func callbackPageHTML(title, message string, success bool) string {
	color := "#ef4444" // red for failure
	icon := "✗"
	if success {
		color = "#8b5cf6" // butea purple for success
		icon = "✓"
	}
	return fmt.Sprintf(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>%s — butea</title>
  <style>
    *{box-sizing:border-box;margin:0;padding:0}
    body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;
         background:#0d0d0f;color:#e8e8ec;min-height:100vh;
         display:flex;align-items:center;justify-content:center}
    .card{background:#1c1c20;border:1px solid #2a2a30;border-radius:16px;
          padding:40px 48px;text-align:center;max-width:400px}
    .icon{font-size:48px;color:%s;margin-bottom:16px}
    h1{font-size:20px;font-weight:600;margin-bottom:8px}
    p{font-size:14px;color:#888;line-height:1.5}
    .badge{display:inline-block;margin-top:20px;padding:6px 14px;
           background:#1a1a2e;border:1px solid #2a2a3a;border-radius:8px;
           font-size:12px;color:#8b5cf6;font-family:monospace}
  </style>
</head>
<body>
  <div class="card">
    <div class="icon">%s</div>
    <h1>%s</h1>
    <p>%s</p>
    <div class="badge">butea CLI</div>
  </div>
</body>
</html>`, title, color, icon, title, message)
}

