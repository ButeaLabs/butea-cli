# Authentication

## How it works

`butea` uses a **browser-based OAuth flow** — no passwords are ever typed into the terminal.

```
butea init
  │
  ├─ starts a local HTTP server on 127.0.0.1 (random free port)
  ├─ generates a 128-bit random state token (CSRF protection)
  ├─ opens: https://app.butea.app/signin?cli_port=PORT&cli_state=STATE
  │
  └─ browser: user signs in via GitHub / GitLab / Google
       │
       └─ /auth/callback detects cli_port in sessionStorage
            │
            └─ redirects to: http://127.0.0.1:PORT/auth/callback
                              ?access_token=AT
                              &refresh_token=RT
                              &state=STATE
                 │
                 └─ CLI validates state, saves tokens, shuts down server
```

The browser shows **"✓ You can close this tab"** and the terminal resumes.

---

## Credentials storage

```
~/.butea/cred.toml   (mode 0600 — owner read-only)
```

```toml
access_token  = "eyJ..."
refresh_token = "rt_..."
```

Tokens never appear in shell history, server logs, or environment variables.

---

## Silent token refresh

Every API call that returns `401 Unauthorized`:

1. The CLI calls `POST /auth/refresh` with the stored refresh token
2. New tokens are saved to `~/.butea/cred.toml`
3. The original request is retried once — transparent to the user

If the refresh also fails (e.g. the session was revoked from the dashboard), the error message says:

```
not logged in – run 'butea init' to authenticate
```

---

## Commands

| Command | What it does |
|---|---|
| `butea init` | Browser auth flow + create `~/.butea/` |
| `butea init --reauth` | Force re-authentication even if already logged in |
| `butea auth whoami` | Print the authenticated user's profile |
| `butea auth logout` | Revoke session server-side + clear `~/.butea/cred.toml` |

---

## Security notes

| Concern | How it's handled |
|---|---|
| CSRF | 128-bit random state token validated on callback |
| Token exposure | Tokens travel on `127.0.0.1` loopback only |
| File access | `cred.toml` written at `0600` (no group/world read) |
| Replay | Local server shuts down immediately after one successful callback |

