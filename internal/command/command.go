package command

import (
	"bytes"
)

type Command struct {
	ID      string   `json:"id"`
	Command []string `json:"command"`
}

// Execute runs the command with the given config and outputs
func (c Command) Execute(opts *Config, outputs *Outputs) (stdout *bytes.Buffer, stderr *bytes.Buffer, err error) {
	// TODO: run the command, return the output
	return &bytes.Buffer{}, &bytes.Buffer{}, nil
}
