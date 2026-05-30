# butea-cli

> The official command-line interface for the [Butea](https://app.butea.app) platform.  
> Deploy projects, manage deployments and more — without leaving your terminal or IDE.

[![npm version](https://img.shields.io/npm/v/butea-cli)](https://www.npmjs.com/package/butea-cli)

---

## Install

```bash
npm install -g butea-cli
```

Requires **Node.js 18+**. The correct binary for your OS/architecture is installed automatically.

---

## Quick start

```bash
# 1. Authenticate (opens browser)
butea init

# 2. Link a repo to a project
cd my-app
butea init          # → creates .butea.toml

# 3. Deploy
butea deploy

# 4. Check status
butea deployments list <projectId>
butea health
```

---

## Commands

| Command | Description |
|---|---|
| `butea init` | Authenticate + optionally link a project |
| `butea deploy [id] [-b branch]` | Trigger a deployment |
| `butea projects list / get / delete` | Manage projects |
| `butea deployments list / get / cancel` | Manage deployments |
| `butea auth whoami` | Show authenticated user |
| `butea auth logout` | Sign out |
| `butea health` | Check API status |

Full reference → [docs/commands.md](docs/commands.md)

---

## Configuration

```
~/.butea/
├── config.toml    # api_url, app_url
└── cred.toml      # tokens  [0600]

<repo>/
└── .butea.toml    # project_id, branch
```

Override via env vars for CI/CD:

```bash
BUTEA_API_URL=https://api.mycompany.com
BUTEA_APP_URL=https://app.mycompany.com
```

Full reference → [docs/configuration.md](docs/configuration.md)

---

## Documentation

| Doc | Contents |
|---|---|
| [docs/getting-started.md](docs/getting-started.md) | Installation, first deploy, common workflows |
| [docs/authentication.md](docs/authentication.md) | Browser OAuth flow, token refresh, security model |
| [docs/configuration.md](docs/configuration.md) | Config files, env vars, resolution order |
| [docs/commands.md](docs/commands.md) | All commands with flags and example output |
| [docs/contributing.md](docs/contributing.md) | Local build, tests, adding commands, publishing |

---

## How authentication works

`butea init` opens your browser — you sign in once via GitHub, GitLab, or Google.  
Tokens are stored in `~/.butea/cred.toml` (mode `0600`) and silently refreshed in the background.  
Nothing is ever typed into the terminal.

Details → [docs/authentication.md](docs/authentication.md)

---

## License

MIT — see [LICENSE](LICENSE)
