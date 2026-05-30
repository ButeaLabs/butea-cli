package main

import "github.com/ButeaLabs/butea-cli/cmd"

// version is injected at build time via:
//
//	go build -ldflags "-X main.version=0.1.0"
//
// It falls back to "dev" for local builds.
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
