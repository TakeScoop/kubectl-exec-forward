package command

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"text/template"

	"github.com/tidwall/gjson"
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
	name, args, err := c.render(config, cmdArgs, outputs, true)
	if err != nil {
		return nil, err
	}

	// nolint:gosec
	return exec.CommandContext(ctx, name, args...), nil
}

func (c Command) render(config *Config, cmdArgs *Args, outputs map[string]Output, showSensitive bool) (name string, args []string, err error) {
	name = c.Command[0]
	rawArgs := []string{}

	if len(c.Command) > 1 {
		rawArgs = append(rawArgs, c.Command[1:]...)
	}

	args = make([]string, len(rawArgs))

	for i, a := range rawArgs {
		tpl, err := template.New(c.ID).Option("missingkey=error").Funcs(template.FuncMap{
			"trim":      strings.TrimSpace,
			"json":      gjson.Get,
			"sensitive": sensitiveFunc(showSensitive),
		}).Parse(a)
		if err != nil {
			return "", nil, err
		}

		o := new(bytes.Buffer)

		if err := tpl.Execute(o, &templateInputs{
			Config:  config,
			Args:    cmdArgs,
			Outputs: outputs,
		}); err != nil {
			return "", nil, err
		}

		args[i] = o.String()
	}

	return name, args, nil
}

// execute runs the command with the given config and outputs.
func (c Command) execute(ctx context.Context, config *Config, args *Args, previousOutputs map[string]Output, streams *genericclioptions.IOStreams) (map[string]Output, error) {
	outputs := map[string]Output{}

	for k, v := range previousOutputs {
		outputs[k] = v
	}

	cmd, err := c.toCmd(ctx, config, args, outputs)
	if err != nil {
		return nil, err
	}

	bout := new(bytes.Buffer)
	berr := new(bytes.Buffer)

	ows := []io.Writer{bout}
	ews := []io.Writer{berr}

	if c.Interactive || config.Verbose {
		ows = append(ows, streams.Out)
		ews = append(ews, streams.ErrOut)
	}

	if c.Interactive {
		cmd.Stdin = streams.In
	}

	cmd.Stdout = io.MultiWriter(ows...)
	cmd.Stderr = io.MultiWriter(ews...)

	if err := cmd.Run(); err != nil {
		name, args, _ := c.render(config, args, outputs, false)
		args = append([]string{name}, args...)

		fmt.Fprintf(streams.ErrOut, "Error running command: %v\n", args)
		fmt.Fprintf(streams.ErrOut, "%s\n", berr)

		return nil, err
	}

	if c.ID != "" {
		outputs[c.ID] = Output{
			Stdout: bout.String(),
			Stderr: berr.String(),
		}
	}

	return outputs, err
}

// parseCommand returns a Command from an annotation storing a single command in json format.
func parseComand(annotations map[string]string, key string) (command Command, err error) {
	v, ok := annotations[key]
	if !ok {
		return command, nil
	}

	if err := json.Unmarshal([]byte(v), &command); err != nil {
		return command, err
	}

	return command, nil
}
