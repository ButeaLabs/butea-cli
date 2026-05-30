# Getting Started

## Installation

```bash
npm install -g butea-cli
```

Verify the install:

```bash
butea --version
butea health
```

---

## First-time setup

Run `butea init` in any directory:

```bash
butea init
```

What happens:
1. Your browser opens to `https://butea.in/signin`
2. Sign in with GitHub, GitLab, or Google
3. The browser hands tokens back to the CLI automatically
4. `~/.butea/` is created with your credentials
5. You are optionally asked whether to link the current directory to a project

---

## Linking a repository

Inside a project you want to deploy from:

```bash
cd my-app
butea init
# → "Link this directory to a Butea project? [y/N]" → y
# → pick from numbered list or paste a project ID
# → .butea.toml is created
```

Once `.butea.toml` exists you can deploy with zero flags:

```bash
butea deploy
```

Commit `.butea.toml` to share the project linkage with your team, or add it to `.gitignore` to keep it personal.

---

## Deploying

```bash
# uses .butea.toml (project_id + branch)
butea deploy

# explicit project, default branch
butea deploy <projectId>

# explicit project + branch
butea deploy <projectId> --branch staging
```

---

## Checking your session

```bash
butea auth whoami      # who am I?
butea health           # is the API reachable?
```

---

## Signing out

```bash
butea auth logout
```

Tokens are revoked server-side and `~/.butea/cred.toml` is cleared.

