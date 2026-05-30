package main

import "github.com/ButeaLabs/butea-cli/cmd"

// version is injected at build time via:
//
//	go build -ldflags "-X github.com/ButeaLabs/butea-cli/cmd.Version=v1.2.3"
//
// It falls back to "dev" for local builds.
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
