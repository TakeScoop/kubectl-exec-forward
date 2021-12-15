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

type commandOptions struct {
	config  *Config
	args    *Args
	outputs map[string]Output
}

// toCmd returns a golang cmd object from the calling command.
func (c Command) toCmd(ctx context.Context, commandOpts *commandOptions) (*exec.Cmd, error) {
	name := c.Command[0]
	rawArgs := []string{}

	if len(c.Command) > 1 {
		rawArgs = append(rawArgs, c.Command[1:]...)
	}

	args := make([]string, len(rawArgs))

	for i, a := range rawArgs {
		tpl := template.Must(template.New(c.ID).Option("missingkey=error").Funcs(template.FuncMap{
			"trim": strings.TrimSpace,
		}).Parse(a))

		o := new(bytes.Buffer)

		if err := tpl.Execute(o, commandOpts.toInterface()); err != nil {
			return nil, err
		}

		args[i] = o.String()
	}

	// nolint:gosec
	return exec.CommandContext(ctx, name, args...), nil
}

// toInterface takes a CommandOptions object and returns a generic interface map for usage within a tempate execution.
func (co commandOptions) toInterface() map[string]interface{} {
	input := map[string]interface{}{}

	input["Config"] = co.config
	input["Args"] = co.args
	input["Outputs"] = map[string]Output{}

	for k, v := range co.outputs {
		input["Outputs"].(map[string]Output)[k] = v
	}

	return input
}

// execute runs the command with the given config and outputs.
func (c Command) execute(ctx context.Context, opts *commandOptions, streams *genericclioptions.IOStreams) (Output, error) {
	cmd, err := c.toCmd(ctx, opts)
	if err != nil {
		return Output{}, err
	}

	bout := new(bytes.Buffer)
	berr := new(bytes.Buffer)

	ows := []io.Writer{bout}
	ews := []io.Writer{berr}

	if c.Interactive || opts.config.Verbose {
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
