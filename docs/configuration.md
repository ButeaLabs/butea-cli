# Configuration

## File layout

```
~/.butea/
├── config.toml     global settings     (mode 0644)
└── cred.toml       auth tokens         (mode 0600)

<repo>/
└── .butea.toml     per-project config  (mode 0644)
```

---

## `~/.butea/config.toml` — global settings

Created automatically by `butea init`.

```toml
api_url = "https://api.butea.app"
app_url = "https://app.butea.app"
```

| Key | Purpose |
|---|---|
| `api_url` | Backend REST API (Go / sal) — all CLI HTTP calls go here |
| `app_url` | Frontend web app (SvelteKit) — browser opens here for auth |

---

## `~/.butea/cred.toml` — credentials

```toml
access_token  = "eyJ..."
refresh_token = "rt_..."
```

Written at `0600` (owner-read-only). Never commit or share this file.

---

## `.butea.toml` — per-repository project config

Created by `butea init` (optional step) or manually.

```toml
project_id = "abc-123-uuid"
branch     = "main"
```

`butea deploy` reads this file when no project ID is passed as an argument.

You can commit this file so the whole team shares the project linkage, or add it to `.gitignore` to keep it personal.

---

## Resolution order

Settings are resolved in this priority (highest first):

| Source | Example |
|---|---|
| CLI flags | `butea deploy --api-url https://api.company.com` |
| Environment variables | `BUTEA_API_URL=…` `BUTEA_APP_URL=…` |
| `~/.butea/config.toml` | saved by `butea init` |
| Built-in defaults | `api.butea.app` / `app.butea.app` |

---

## Environment variables

Useful for CI/CD pipelines, Docker, or self-hosted deployments where you don't want to write a config file.

| Variable | Overrides |
|---|---|
| `BUTEA_API_URL` | `api_url` in `config.toml` |
| `BUTEA_APP_URL` | `app_url` in `config.toml` |

### Example — GitHub Actions

```yaml
env:
  BUTEA_API_URL: https://api.mycompany.internal
  BUTEA_APP_URL: https://app.mycompany.internal
```

### Example — Docker

```dockerfile
ENV BUTEA_API_URL=https://api.mycompany.internal
ENV BUTEA_APP_URL=https://app.mycompany.internal
```

---

## CLI flags (per-invocation override)

```bash
butea --api-url http://localhost:8080 health
butea --api-url http://localhost:8080 --app-url http://localhost:5173 init
```

Flags override everything else for that single command only — nothing is saved.

