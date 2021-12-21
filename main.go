package main

import (
	"github.com/takescoop/kubectl-port-forward-hooks/cmd"
)

// Version will be replaced with the Git tag version at build time during release.
// nolint: gochecknoglobals
var Version string

func main() {
	cmd.Execute(Version)
}
