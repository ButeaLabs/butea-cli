# butea-cli — Implementation Guide

> **Date:** May 30, 2026  
> **Module:** `github.com/ButeaLabs/butea-cli`  
> **npm package:** `@butea/cli`  
> **Install:** `npm install -g @butea/cli`

---

## Overview

`butea-cli` is the official command-line interface for the Butea platform. It is:
- Written in **Go** (compiled, no runtime dependency)
- Shipped via **npm** using the optional-dependency binary pattern (same as `esbuild`, `@biomejs/biome`)
- Authenticated via a **browser-based OAuth flow** — no passwords typed into the terminal
- Configured via **TOML files** in well-known OS locations

---

## File layout

```
butea-cli/
├── main.go                           Entry point → cmd.Execute()
├── go.mod / go.sum                   cobra · x/term · BurntSushi/toml
├── Makefile                          build / test / release / publish targets
│
├── cmd/
│   ├── root.go                       Root command, --api-url flag, loadAll() + newClient()
│   ├── init.go          ★            butea init  (browser auth + project linking)
│   ├── auth.go                       butea auth logout / whoami
│   ├── deploy.go                     butea deploy  (reads .butea.toml)
│   ├── projects.go                   butea projects list / get / delete
│   ├── deployments.go                butea deployments list / get / cancel
│   └── health.go                     butea health
│
├── internal/
│   ├── auth/
│   │   └── flow.go      ★            Browser OAuth handshake (local callback server)
│   ├── api/
│   │   ├── client.go                 Typed HTTP client + silent token refresh
│   │   ├── types.go                  Structs mirroring sal BE DTOs
│   │   └── client_test.go            13 httptest-based tests
│   └── config/
│       ├── config.go    ★            TOML config: GlobalConfig · Credentials · LocalConfig
│       └── config_test.go            12 tests
│
└── npm/
    └── @butea/
        ├── cli/                      Main npm package (name: @butea/cli)
        │   ├── package.json
        │   └── bin/butea.js          Node shim — resolves platform binary
        ├── cli-darwin-arm64/
        ├── cli-darwin-x64/
        ├── cli-linux-arm64/
        ├── cli-linux-x64/
        ├── cli-windows-arm64/
        └── cli-windows-x64/
```

---

## Configuration files

```
~/.butea/
├── config.toml        # global: api_url, app_url
└── cred.toml          # credentials: access_token, refresh_token  [mode 0600]

<repo>/
└── .butea.toml        # per-project: project_id, branch
```

### `~/.butea/config.toml`
```toml
api_url = "https://api.butea.app"
app_url = "https://app.butea.app"
```

### `~/.butea/cred.toml` (mode 0600, never shown in logs)
```toml
access_token  = "eyJ..."
refresh_token = "rt_..."
```

### `.butea.toml` (per-repo, optional)
```toml
project_id = "abc-123-uuid"
branch     = "main"
```
Add to `.gitignore` or commit it — either approach works.

---

## Commands

### `butea init` ★ (the main entry point)

```
butea init              # first-time setup or re-link a project
butea init --reauth     # force sign in even if already authenticated
butea init --link       # skip the y/N prompt and always link a project
```

**What it does:**
1. Opens the browser to `https://app.butea.app/signin?cli_port=PORT&cli_state=STATE`
2. Prints the URL to the terminal in case the browser doesn't open automatically
3. Waits for the local callback server (5-minute timeout)
4. After the user signs in, the browser redirects to `http://127.0.0.1:PORT/auth/callback?...`
5. CLI validates the state token (CSRF protection), saves tokens
6. Creates `~/.butea/config.toml` and `~/.butea/cred.toml`
7. Optionally shows the project list and creates `.butea.toml` in the current directory

### Authentication commands

| Command | Description | BE endpoint |
|---|---|---|
| `butea auth logout` | Revoke session and clear `~/.butea/cred.toml` | `POST /auth/logout` |
| `butea auth whoami` | Print authenticated user's profile | `GET /api/v1/me` |

### Projects

| Command | Description | BE endpoint |
|---|---|---|
| `butea projects list` | Table of all projects | `GET /api/v1/projects/` |
| `butea projects get <id>` | Detailed project view | `GET /api/v1/projects/{id}` |
| `butea projects delete <id>` | Delete (prompts for confirmation) | `DELETE /api/v1/projects/{id}` |

### Deployments

| Command | Description | BE endpoint |
|---|---|---|
| `butea deploy` | Trigger deployment (reads `.butea.toml`) | `POST /api/v1/projects/{id}/deployments` |
| `butea deploy <id>` | Trigger deployment for explicit project | same |
| `butea deploy <id> --branch feat` | Deploy a specific branch | same |
| `butea deployments list <projectId>` | Paginated deployment list | `GET /api/v1/projects/{id}/deployments` |
| `butea deployments get <id>` | Single deployment details | `GET /api/v1/deployments/{id}` |
| `butea deployments cancel <id>` | Cancel queued/pending deployment | `POST /api/v1/deployments/{id}/cancel` |

### Other

| Command | Description |
|---|---|
| `butea health` | Check API reachability + service statuses |
| `butea --api-url URL <command>` | Override API URL for one invocation |

---

## Browser OAuth flow (detail)

```
CLI                                    Browser                        BE / FE
 │                                        │                             │
 │  start local server (random port P)   │                             │
 │  generate state token S               │                             │
 │  open browser ──────────────────────► │                             │
 │                                        │  GET /signin?cli_port=P    │
 │                                        │    &cli_state=S            │
 │                                        │  store P+S in sessionStorage│
 │                                        │  user signs in (OAuth/pw)  │
 │                                        │                  ──────────►│
 │                                        │                  ◄──────────│
 │                                        │  GET /auth/callback        │
 │                                        │    #access_token=AT        │
 │                                        │    &refresh_token=RT       │
 │                                        │  detect cli_port in ss     │
 │                                        │  redirect to local server  │
 │  GET /auth/callback?access_token=AT ◄──│                             │
 │       &refresh_token=RT&state=S        │                             │
 │  validate state == S                   │                             │
 │  save tokens to ~/.butea/cred.toml    │                             │
 │  show success HTML page ─────────────► │  "✓ You can close this tab"│
 │  shutdown local server                 │                             │
```

**Security notes:**
- State token is 16 random bytes (128-bit) — prevents CSRF
- Local server binds to `127.0.0.1` only (not `0.0.0.0`)
- Tokens travel on loopback only — never leave the machine until saved
- `cred.toml` written with mode `0600` (owner-read-only)

---

## FE changes made (aparajita)

### `src/routes/signin/+page.svelte`
- On `onMount`, reads `?cli_port` and `?cli_state` from URL
- If present, stores them in `sessionStorage`
- sessionStorage survives the OAuth redirect chain (same browser tab/origin)

### `src/routes/auth/callback/+page.svelte`
- After tokens arrive, checks `sessionStorage` for `cli_port` + `cli_state`
- If present: removes them from sessionStorage, then redirects to  
  `http://127.0.0.1:{cli_port}/auth/callback?access_token=...&refresh_token=...&state=...`
- If absent: normal browser session setup (unchanged behaviour)

---

## Silent token refresh

The API client (`internal/api/client.go`) intercepts any `401 Unauthorized` and:
1. Calls `POST /auth/refresh` with the stored refresh token
2. Saves new tokens via `OnTokenRefresh` callback → `cred.Save()`
3. Retries the original request once

If the refresh also fails, the original error surfaces and the user is told to run `butea init` again.

---

## Testing

```bash
go test ./...          # 25 tests across api + config packages
make test-verbose      # verbose output
```

### API client tests (13) — `internal/api/client_test.go`
Uses `net/http/httptest` — real HTTP round-trips, no mocking frameworks.

| Test | What it covers |
|---|---|
| `TestLogin_Success` | Token extraction from response |
| `TestLogin_BadCredentials` | 401 → error propagation |
| `TestGetMe_Success` | Authenticated GET |
| `TestGetMe_Unauthorized` | Persistent 401 (refresh also fails) |
| `TestAutoRefresh_RetriesAfter401` | Refresh → retry → success; `OnTokenRefresh` called |
| `TestListProjects_Success` | Project list parsing |
| `TestCreateDeployment_Success` | Deployment creation, 201 |
| `TestHealth_Healthy` | Health JSON parsing |
| `TestHealth_Degraded` | 503 → APIError surfaced |
| `TestAPIError_Format` | `"CODE: message"` format |
| `TestAPIError_NoCode` | Plain message when code absent |
| `TestCancelDeployment_Success` | 204 No Content |
| `TestDeleteProject_Success` | DELETE + 204 |

### Config tests (12) — `internal/config/config_test.go`
Uses `t.TempDir()` + `t.Setenv("HOME", ...)` — never touches real `~/.butea/`.

| Test | What it covers |
|---|---|
| `TestLoadGlobal_DefaultsWhenMissing` | First-run defaults |
| `TestGlobalConfig_SaveAndLoad` | TOML round-trip for GlobalConfig |
| `TestLoadCredentials_EmptyWhenMissing` | Empty cred on first run |
| `TestCredentials_SaveAndLoad` | TOML round-trip for Credentials |
| `TestCredentials_FilePermissions` | cred.toml must be mode 0600 |
| `TestCredentials_Clear` | Clear zeroes tokens |
| `TestCredentials_RequireAuth` | Error with no token |
| `TestCredentials_IsLoggedIn` | Boolean helper |
| `TestGlobalDir_IsUnderHome` | ~/.butea path |
| `TestGlobalConfig_CreatesDirectory` | Directory auto-created on Save |
| `TestLoadLocal_NilWhenMissing` | nil when no .butea.toml |
| `TestLocalConfig_SaveAndLoad` | TOML round-trip for LocalConfig |

---

## Dependencies

| Package | Version | Why |
|---|---|---|
| `github.com/spf13/cobra` | v1.10.2 | CLI framework (pre-existing) |
| `golang.org/x/term` | v0.43.0 | Suppress echo for password fallback |
| `github.com/BurntSushi/toml` | v1.6.0 | TOML config file parsing/writing |

---

## npm packaging

### Install for end users
```bash
npm install -g @butea/cli
butea init
```

### Publish a new version
```bash
# 1. Build all platform binaries
make release

# 2. Publish all packages
make publish
```

The `make release` target cross-compiles for all 6 platforms. Binaries land in `npm/@butea/cli-{platform}/bin/`. The main `@butea/cli` package's Node shim then resolves the right binary at runtime.

---

## Quick start

```bash
# Install
npm install -g @butea/cli

# First-time setup (opens browser)
butea init

# Inside a project repo — link to your Butea project
cd my-app
butea init    # choose a project when prompted → creates .butea.toml

# Deploy
butea deploy

# Check status
butea deployments list <projectId>
butea health

# Sign out
butea auth logout
```

## Local development

```bash
# Build
go build -o butea .

# Run against a local backend
./butea --api-url http://localhost:8080 health
./butea --api-url http://localhost:8080 init

# Tests
go test ./...
```
