# Contributing & Development

## Prerequisites

- Go 1.22+
- Node.js 18+ (only needed to test the npm shim)

---

## Build

```bash
# Local binary (current platform)
go build -o butea .

# All platforms (lands in npm/@butea/cli-{platform}/bin/)
make release
```

---

## Test

```bash
go test ./...          # all tests
make test-verbose      # with -v flag
```

### Test coverage

| Package | Tests | What they cover |
|---|---|---|
| `internal/api` | 13 | HTTP client, token refresh, error parsing — real `httptest` servers |
| `internal/config` | 16 | TOML save/load, file permissions, env var override |

Tests never touch the real `~/.butea/` directory — they use `t.TempDir()` + `t.Setenv("HOME", …)`.

---

## Project structure

```
butea-cli/
├── main.go                    entry point
├── Makefile                   build / test / release / publish
│
├── cmd/                       one file per Cobra command group
│   ├── root.go                root cmd + shared helpers (loadAll, newClient)
│   ├── init.go                butea init  — browser auth + project link
│   ├── auth.go                butea auth logout / whoami
│   ├── deploy.go              butea deploy
│   ├── projects.go            butea projects list / get / delete
│   ├── deployments.go         butea deployments list / get / cancel
│   └── health.go              butea health
│
├── internal/
│   ├── auth/flow.go           browser OAuth handshake (local callback server)
│   ├── api/
│   │   ├── client.go          typed HTTP client + silent token refresh
│   │   ├── types.go           request/response structs mirroring BE DTOs
│   │   └── client_test.go
│   └── config/
│       ├── config.go          TOML config — GlobalConfig · Credentials · LocalConfig
│       └── config_test.go
│
├── docs/                      this directory
│
└── npm/
    └── @butea/
        ├── cli/               main npm package  (@butea/cli)
        │   └── bin/butea.js   Node shim — resolves platform binary
        └── cli-{platform}/    one package per platform binary
```

---

## Adding a new command

1. Create `cmd/<name>.go`
2. Define a `*cobra.Command` and register it in `init()` via `rootCmd.AddCommand(...)`
3. Use `loadAll()` to get config + credentials, `newClient(cfg, cred)` for the API client
4. Return errors — Cobra prints them cleanly; use `fatal()` only for unrecoverable setup errors

```go
var myCmd = &cobra.Command{
    Use:   "my-command",
    Short: "Does something useful",
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg, cred, err := loadAll()
        if err != nil {
            return err
        }
        if err := cred.RequireAuth(); err != nil {
            return err
        }
        client := newClient(cfg, cred)
        // ... call client methods
        return nil
    },
}

func init() {
    rootCmd.AddCommand(myCmd)
}
```

---

## Adding a new API method

1. Add the request/response types to `internal/api/types.go`
2. Add the method to `internal/api/client.go`
3. Write a test in `internal/api/client_test.go` using `httptest.NewServer`

---

## Publishing to npm

```bash
# 1. Build all platform binaries
make release

# 2. Publish all seven packages (@butea/cli + 6 platform packages)
make publish
```

The publish step requires `npm login` with access to the `@butea` scope.

### Version bump

Update `version` in all `npm/@butea/*/package.json` files and in the `Makefile`'s `VERSION` variable before running `make publish`.

