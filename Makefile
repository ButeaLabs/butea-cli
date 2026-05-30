BINARY_NAME := butea
VERSION     ?= 0.1.0
MODULE      := github.com/ButeaLabs/butea-cli

# -X injects the version into cmd.Version (the exported var read by all commands).
LDFLAGS     := -ldflags "-s -w -X main.version=$(VERSION)"

# npm package root — flat unscoped layout matching the published "butea-cli" package
NPM_DIR     := $(CURDIR)/npm

.PHONY: build test test-verbose clean release dev-install dev-install-x64 publish

# ── Local build ──────────────────────────────────────────────────────────────

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# ── Tests ────────────────────────────────────────────────────────────────────

test:
	go test ./... -count=1

test-verbose:
	go test ./... -v -count=1

# ── Local dev install (no npm publish needed) ─────────────────────────────────
# Builds the native binary for this machine, then npm-links butea-cli globally
# so `butea` works in any terminal immediately.

dev-install:
	GOOS=darwin GOARCH=arm64 go build -trimpath $(LDFLAGS) -o $(NPM_DIR)/butea-cli-darwin-arm64/bin/$(BINARY_NAME) .
	chmod +x $(NPM_DIR)/butea-cli-darwin-arm64/bin/$(BINARY_NAME)
	cd $(NPM_DIR)/butea-cli-darwin-arm64 && npm link
	cd $(NPM_DIR)/butea-cli && npm link butea-cli-darwin-arm64 && npm link
	@echo ""
	@echo "✓  Run: butea --version"

dev-install-x64:
	GOOS=darwin GOARCH=amd64 go build -trimpath $(LDFLAGS) -o $(NPM_DIR)/butea-cli-darwin-x64/bin/$(BINARY_NAME) .
	chmod +x $(NPM_DIR)/butea-cli-darwin-x64/bin/$(BINARY_NAME)
	cd $(NPM_DIR)/butea-cli-darwin-x64 && npm link
	cd $(NPM_DIR)/butea-cli && npm link butea-cli-darwin-x64 && npm link
	@echo "✓  Run: butea --version"

# ── Release: compile for all 6 platforms ─────────────────────────────────────
# Binaries land in npm/butea-cli-{platform}/bin/

release:
	GOOS=darwin  GOARCH=arm64 go build -trimpath $(LDFLAGS) -o $(NPM_DIR)/butea-cli-darwin-arm64/bin/$(BINARY_NAME) .
	GOOS=darwin  GOARCH=amd64 go build -trimpath $(LDFLAGS) -o $(NPM_DIR)/butea-cli-darwin-x64/bin/$(BINARY_NAME) .
	GOOS=linux   GOARCH=arm64 go build -trimpath $(LDFLAGS) -o $(NPM_DIR)/butea-cli-linux-arm64/bin/$(BINARY_NAME) .
	GOOS=linux   GOARCH=amd64 go build -trimpath $(LDFLAGS) -o $(NPM_DIR)/butea-cli-linux-x64/bin/$(BINARY_NAME) .
	GOOS=windows GOARCH=arm64 go build -trimpath $(LDFLAGS) -o $(NPM_DIR)/butea-cli-windows-arm64/bin/$(BINARY_NAME).exe .
	GOOS=windows GOARCH=amd64 go build -trimpath $(LDFLAGS) -o $(NPM_DIR)/butea-cli-windows-x64/bin/$(BINARY_NAME).exe .
	@# Make unix binaries executable
	chmod +x \
		$(NPM_DIR)/butea-cli-darwin-arm64/bin/$(BINARY_NAME) \
		$(NPM_DIR)/butea-cli-darwin-x64/bin/$(BINARY_NAME) \
		$(NPM_DIR)/butea-cli-linux-arm64/bin/$(BINARY_NAME) \
		$(NPM_DIR)/butea-cli-linux-x64/bin/$(BINARY_NAME)
	@echo "✓  All platform binaries built."

# ── Publish (requires: npm login) ────────────────────────────────────────────
# Publish the 6 platform packages first, then the main butea-cli shim.
# Order matters: optionalDependencies must already exist on the registry.

PLATFORM_PKGS := \
	$(NPM_DIR)/butea-cli-darwin-arm64 \
	$(NPM_DIR)/butea-cli-darwin-x64 \
	$(NPM_DIR)/butea-cli-linux-arm64 \
	$(NPM_DIR)/butea-cli-linux-x64 \
	$(NPM_DIR)/butea-cli-windows-arm64 \
	$(NPM_DIR)/butea-cli-windows-x64

publish: release
	@for dir in $(PLATFORM_PKGS); do \
	  echo "→ Publishing $$(basename $$dir) …"; \
	  (cd "$$dir" && npm publish --access public) || exit 1; \
	done
	@echo "→ Publishing butea-cli (main shim) …"
	cd $(NPM_DIR)/butea-cli && npm publish --access public
	@echo ""
	@echo "✓  All packages published."
	@echo "   Users can now run: npm install -g butea-cli"

# ── Clean ─────────────────────────────────────────────────────────────────────

clean:
	rm -f $(BINARY_NAME)
	rm -f \
		$(NPM_DIR)/butea-cli-darwin-arm64/bin/$(BINARY_NAME) \
		$(NPM_DIR)/butea-cli-darwin-x64/bin/$(BINARY_NAME) \
		$(NPM_DIR)/butea-cli-linux-arm64/bin/$(BINARY_NAME) \
		$(NPM_DIR)/butea-cli-linux-x64/bin/$(BINARY_NAME) \
		$(NPM_DIR)/butea-cli-windows-arm64/bin/$(BINARY_NAME).exe \
		$(NPM_DIR)/butea-cli-windows-x64/bin/$(BINARY_NAME).exe
