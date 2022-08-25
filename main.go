// Package main is the entrypoint for the exec-forward CLI.
package main

import (
	"github.com/takescoop/kubectl-exec-forward/cmd"
)

// version will be replaced with the Git tag version at build time during release.
var version = "dev"

func main() {
	cmd.Execute(version)
}
