package command

// Config stores configuration which is used to construct the tunnel as well as passed to the hook commands.
type Config struct {
	LocalPorts      []int
	Verbose         bool
	CommandOverride string
}
