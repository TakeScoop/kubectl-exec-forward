package command

import (
	"bytes"
	"context"
	"encoding/json"
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
	config *Config
	args   *Args
}

// toCmd returns a golang cmd object from the calling command.
func (c Command) toCmd(ctx context.Context, commandOpts *commandOptions) (*exec.Cmd, error) {
	name := c.Command[0]
	rawCmdArgs := []string{}

	opts, err := commandOpts.toInterface()
	if err != nil {
		return nil, err
	}

	if len(c.Command) > 1 {
		rawCmdArgs = append(rawCmdArgs, c.Command[1:]...)
	}

	cmdArgs := make([]string, len(rawCmdArgs))

	for i, a := range rawCmdArgs {
		tpl := template.Must(template.New(c.ID).Funcs(template.FuncMap{
			"trim": strings.TrimSpace,
		}).Parse(a))

		o := new(bytes.Buffer)

		if err := tpl.Execute(o, opts); err != nil {
			return nil, err
		}

		cmdArgs[i] = o.String()
	}

	// nolint:gosec
	return exec.CommandContext(ctx, name, cmdArgs...), nil
}

// toInterface takes a CommandOptions object and returns a generic interface map for usage within a tempate execution.
func (co commandOptions) toInterface() (map[string]interface{}, error) {
	input := map[string]interface{}{}

	c, err := json.Marshal(co.config)
	if err != nil {
		return nil, err
	}

	i := map[string]interface{}{}
	if err := json.Unmarshal(c, &i); err != nil {
		return nil, err
	}

	input["Config"] = i
	input["Args"] = co.args

	return input, nil
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
