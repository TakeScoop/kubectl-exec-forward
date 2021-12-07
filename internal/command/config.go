package command

import "time"

// Config stores configuration which is used to construct the tunnel as well as passed to the hook commands.
type Config struct {
	LocalPort  int
	Verbose    bool
	PodTimeout time.Duration
}
