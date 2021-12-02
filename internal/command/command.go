package command

import (
	"bytes"
	"context"
	"io"
	"os/exec"
)

type Command struct {
	ID          string   `json:"id"`
	Command     []string `json:"command"`
	Interactive bool     `json:"interactive"`
}

// Execute runs the command with the given config and outputs
func (c Command) Execute(ctx context.Context, config *Config, arguments *Args, outputs map[string]Output, ios IO) (Output, error) {
	name, args := c.Command[0], c.Command[1:]
	cmd := exec.CommandContext(ctx, name, args...)

	bout := new(bytes.Buffer)
	berr := new(bytes.Buffer)

	ows := []io.Writer{bout}
	ews := []io.Writer{berr}

	if c.Interactive || config.Verbose {
		ows = append(ows, ios.Stdout)
		ews = append(ews, ios.Stderr)
	}

	cmd.Stdout = io.MultiWriter(ows...)
	cmd.Stderr = io.MultiWriter(ews...)
	cmd.Stdin = ios.Stdin

	err := cmd.Run()

	return Output{
		Stdout: bout.String(),
		Stderr: berr.String(),
	}, err
}
