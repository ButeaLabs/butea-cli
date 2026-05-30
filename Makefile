BINARY_NAME := butea
VERSION     ?= 0.1.2
MODULE      := github.com/ButeaLabs/butea-cli

# -X injects the version into cmd.Version (the exported var read by all commands).
LDFLAGS := -ldflags "-s -w -X $(MODULE)/cmd.Version=$(VERSION)"

.PHONY: build test clean release help

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build          Build the CLI binary for the current platform"
	@echo "  test           Run tests"
	@echo "  test-verbose   Run tests with verbose output"
	@echo "  release        Cross-compile binaries for all platforms"
	@echo "  publish        Build all platforms and publish to npm"
	@echo "  clean          Remove built binaries"

# ── Local build ──────────────────────────────────────────────────────────────

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# ── Tests ────────────────────────────────────────────────────────────────────

test:
	go test ./... -count=1

test-verbose:
	go test ./... -v -count=1

# ── Release: compile for all platforms ───────────────────────────────────────
# Binaries land in the matching npm/butea-cli-{platform}/bin/ directory.

release: \
	npm/butea-cli-darwin-arm64/bin/butea \
	npm/butea-cli-darwin-x64/bin/butea \
	npm/butea-cli-linux-arm64/bin/butea \
	npm/butea-cli-linux-x64/bin/butea \
	npm/butea-cli-windows-arm64/bin/butea.exe \
	npm/butea-cli-windows-x64/bin/butea.exe

npm/butea-cli-darwin-arm64/bin/butea:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $@ .

npm/butea-cli-darwin-x64/bin/butea:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $@ .

npm/butea-cli-linux-arm64/bin/butea:
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $@ .

npm/butea-cli-linux-x64/bin/butea:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $@ .

npm/butea-cli-windows-arm64/bin/butea.exe:
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $@ .

npm/butea-cli-windows-x64/bin/butea.exe:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $@ .

# ── Publish (requires npm login) ─────────────────────────────────────────────

publish: release
	@for dir in npm/butea-cli-darwin-arm64 npm/butea-cli-darwin-x64 \
	             npm/butea-cli-linux-arm64 npm/butea-cli-linux-x64 \
	             npm/butea-cli-windows-arm64 npm/butea-cli-windows-x64; do \
	  echo "Publishing $$dir …"; \
	  (cd $$dir && npm publish --access public); \
	done
	@echo "Publishing npm/butea-cli …"
	cp README.md npm/butea-cli/README.md
	(cd npm/butea-cli && npm publish --access public)

clean:
	rm -f $(BINARY_NAME)
	rm -f npm/butea-cli-*/bin/butea npm/butea-cli-*/bin/butea.exe
