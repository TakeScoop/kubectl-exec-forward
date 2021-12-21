package main

import (
	"github.com/takescoop/kubectl-port-forward-hooks/cmd"
)

// version will be replaced with the Git tag version at build time during release.
var version = "dev"

func main() {
	cmd.Execute(version)
}
