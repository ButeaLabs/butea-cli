# Commands Reference

## Global flags

These flags work with every command:

| Flag | Description |
|---|---|
| `--api-url <url>` | Override the backend API URL |
| `--app-url <url>` | Override the frontend app URL |
| `--help` | Show help for any command |

---

## `butea init`

Authenticate and set up butea in the current project.

```
butea init [flags]
```

| Flag | Description |
|---|---|
| `--reauth` | Force re-authentication even if already logged in |
| `--link` | Skip the y/N prompt and always link a project |

**Steps performed:**
1. Opens browser → user signs in → tokens saved to `~/.butea/cred.toml`
2. Creates `~/.butea/config.toml` with default URLs
3. *(Optional)* Lists your projects → you pick one → creates `.butea.toml`

---

## `butea auth`

### `butea auth whoami`

Print the profile of the currently signed-in user.

```bash
butea auth whoami
```

Output:
```
ID:     uuid-...
Name:   Ada Lovelace
Email:  ada@example.com
Active: true
OAuth:  github
Since:  Jan 1 2026
```

### `butea auth logout`

Revoke the current session and clear credentials.

```bash
butea auth logout
```

---

## `butea deploy`

Trigger a new deployment.

```
butea deploy [projectId] [flags]
```

| Flag | Default | Description |
|---|---|---|
| `--branch`, `-b` | project default branch | Branch to deploy |

**Resolution order for project ID:**
1. Positional argument `butea deploy <id>`
2. `project_id` in `.butea.toml` (current directory)
3. Error — must provide one of the above

**Resolution order for branch:**
1. `--branch` flag
2. `branch` in `.butea.toml`
3. Project's `default_branch` (fetched from API)

```bash
# Simplest — reads .butea.toml
butea deploy

# Explicit project
butea deploy abc-123-uuid

# Different branch
butea deploy abc-123-uuid --branch staging
```

---

## `butea deployments`

### `butea deployments list <projectId>`

List paginated deployments for a project.

```bash
butea deployments list abc-123-uuid
```

Output:
```
ID         BRANCH   STATUS    COMMIT    CREATED
──         ──────   ──────    ──────    ───────
dep-uuid   main     success   a1b2c3d   May 30 14:22
dep-uuid   main     queued    f4e5d6c   May 30 14:18
```

### `butea deployments get <deploymentId>`

Show full details for a single deployment.

```bash
butea deployments get dep-uuid
```

### `butea deployments cancel <deploymentId>`

Cancel a deployment that is still queued or pending.

```bash
butea deployments cancel dep-uuid
```

---

## `butea projects`

### `butea projects list`

List all projects visible to the authenticated user.

```bash
butea projects list
# alias: butea projects ls
```

### `butea projects get <projectId>`

Show detailed info for a single project.

```bash
butea projects get abc-123-uuid
```

### `butea projects delete <projectId>`

Delete a project. Prompts for confirmation.

```bash
butea projects delete abc-123-uuid
butea projects delete abc-123-uuid --yes    # skip prompt
# alias: butea projects rm
```

---

## `butea health`

Check API reachability and service statuses.

```bash
butea health
```

Output:
```
✓ API status: healthy (v1.0.0)

Endpoint: https://api.butea.app
Checked:  May 30, 2026 14:22:01 UTC

Services:
  ✓ postgreSQL: healthy (1ms)
  ✓ redis: healthy (0ms)
```

Returns `⚠` and degraded service info when the API is partially down.

