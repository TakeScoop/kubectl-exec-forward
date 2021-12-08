package hooks

import (
	"bytes"
	"context"
	"io"
	"os/exec"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Command represents a runnable command.
type Command struct {
	ID          string   `json:"id"`
	Command     []string `json:"command"`
	Interactive bool     `json:"interactive"`
}

// toCmd returns a golang cmd object from the calling command.
func (c Command) toCmd(ctx context.Context) *exec.Cmd {
	name := c.Command[0]
	args := []string{}

	if len(c.Command) > 1 {
		args = append(args, c.Command[1:]...)
	}

	// nolint:gosec
	return exec.CommandContext(ctx, name, args...)
}

// execute runs the command with the given config and outputs.
// nolint:unparam
func (c Command) execute(ctx context.Context, config *Config, arguments *Args, outputs map[string]Output, streams *genericclioptions.IOStreams) (Output, error) {
	// nolint:godox
	// TODO: Add in go templating to pair the args and config with the passed commands
	cmd := c.toCmd(ctx)

	bout := new(bytes.Buffer)
	berr := new(bytes.Buffer)

	ows := []io.Writer{bout}
	ews := []io.Writer{berr}

	if c.Interactive || config.Verbose {
		ows = append(ows, streams.Out)
		ews = append(ews, streams.ErrOut)
	}

	cmd.Stdout = io.MultiWriter(ows...)
	cmd.Stderr = io.MultiWriter(ews...)
	cmd.Stdin = streams.In

	err := cmd.Run()

	return Output{
		Stdout: bout.String(),
		Stderr: berr.String(),
	}, err
}
