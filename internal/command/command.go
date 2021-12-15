package command

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strings"
	"text/template"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Command represents a runnable command.
type Command struct {
	ID          string   `json:"id"`
	Command     []string `json:"command"`
	Interactive bool     `json:"interactive"`
}

type templateInputs struct {
	Config  *Config
	Args    *Args
	Outputs map[string]Output
}

// toCmd returns a golang cmd object from the calling command.
func (c Command) toCmd(ctx context.Context, config *Config, cmdArgs *Args, outputs map[string]Output) (*exec.Cmd, error) {
	name := c.Command[0]
	rawArgs := []string{}

	if len(c.Command) > 1 {
		rawArgs = append(rawArgs, c.Command[1:]...)
	}

	args := make([]string, len(rawArgs))

	for i, a := range rawArgs {
		tpl, err := template.New(c.ID).Option("missingkey=error").Funcs(template.FuncMap{
			"trim": strings.TrimSpace,
		}).Parse(a)
		if err != nil {
			return nil, err
		}

		o := new(bytes.Buffer)

		if err := tpl.Execute(o, &templateInputs{
			Config:  config,
			Args:    cmdArgs,
			Outputs: outputs,
		}); err != nil {
			return nil, err
		}

		args[i] = o.String()
	}

	// nolint:gosec
	return exec.CommandContext(ctx, name, args...), nil
}

// execute runs the command with the given config and outputs.
func (c Command) execute(ctx context.Context, config *Config, args *Args, outputs map[string]Output, streams *genericclioptions.IOStreams) (Output, error) {
	cmd, err := c.toCmd(ctx, config, args, outputs)
	if err != nil {
		return Output{}, err
	}

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

	if err := cmd.Run(); err != nil {
		return Output{}, err
	}

	return Output{
		Stdout: bout.String(),
		Stderr: berr.String(),
	}, err
}
