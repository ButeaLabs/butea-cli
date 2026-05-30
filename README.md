# butea-cli

> The official command-line interface for the [Butea](https://app.butea.in) platform.  
> Deploy projects, manage deployments and more — without leaving your terminal.

## Install

```bash
npm install -g butea-cli
```

Requires **Node.js 18+**. The correct binary for your OS/architecture is installed automatically.

## Quick start

```bash
# 1. Authenticate and link a project (opens browser)
butea init

# 2. Deploy
butea deploy

# 3. Check status
butea deployments list <projectId>
```

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

## Documentation

Full reference at [docs.butea.in/cli](https://docs.butea.in/cli)

## License

MIT
