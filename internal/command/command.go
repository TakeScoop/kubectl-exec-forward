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
	"github.com/ttacon/chalk"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type commandInput struct {
	ID          string   `json:"id"`
	Command     []string `json:"command"`
	Interactive bool     `json:"interactive"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
}

// Command represents a runnable command.
type Command struct {
	ID          string
	Command     []string
	Interactive bool
	Description string
	hookType    string
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
			"json": gjson.Get,
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

// String returns the command as a human readable string.
func (c Command) String() string {
	str := []string{
		chalk.Yellow.Color(c.hookType),
	}

	if c.ID != "" {
		str = append(str, chalk.Cyan.Color(fmt.Sprintf("(id=%s)", c.ID)))
	}

	if c.Description != "" {
		str = append(str, chalk.Green.Color(c.Description))
	}

	if c.ID == "" && c.Description == "" {
		str = append(str, chalk.Green.Color("no info"))
	}

	return fmt.Sprintf("%s\n", strings.Join(str, ": "))
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

	fmt.Fprintf(streams.ErrOut, "running %s", c)

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
		fmt.Fprint(streams.ErrOut, chalk.Red.Color(fmt.Sprintf("Error running command: %v\n", cmd.Args)))
		fmt.Fprint(streams.ErrOut, chalk.Red.Color(fmt.Sprintf("%s\n", berr)))

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

func (ci commandInput) toCommand(hookType string) Command {
	return Command{
		ID:          ci.ID,
		Command:     ci.Command,
		Interactive: ci.Interactive,
		hookType:    hookType,
		Description: ci.Description,
	}
}

// Returns the hook type based on the passed annotation key.
func getHookTypeFromAnnotationKey(key string) string {
	switch key {
	case PreAnnotation:
		return preConnectHookType
	case PostAnnotation:
		return postConnectHookType
	case CommandAnnotation:
		return commandHookType
	default:
		return ""
	}
}

// parseCommand returns a Command from an annotation storing a single command in json format.
func parseComand(annotations map[string]string, key string) (command Command, err error) {
	v, ok := annotations[key]
	if !ok {
		return command, nil
	}

	var ci commandInput

	if err := json.Unmarshal([]byte(v), &ci); err != nil {
		return command, err
	}

	return ci.toCommand(getHookTypeFromAnnotationKey(key)), nil
}
