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

// Command represents a runnable command.
type Command struct {
	ID          string   `json:"id"`
	Command     []string `json:"command"`
	Interactive bool     `json:"interactive"`
	Name        string   `json:"name"`
}

type TemplateData struct {
	LocalPort int
	Args      Args
	Outputs   map[string]string
}

type TemplateOptions struct {
	ShowSensitive bool
}

// ToCmd returns a Cmd object that can be used with the exec package.
func (c Command) ToCmd(ctx context.Context, data TemplateData) (*exec.Cmd, error) {
	name, args, err := c.Render(data, TemplateOptions{
		ShowSensitive: true,
	})

	if err != nil {
		return nil, err
	}

	// nolint:gosec
	return exec.CommandContext(ctx, name, args...), nil
}

func (c Command) Render(data TemplateData, options TemplateOptions) (name string, args []string, err error) {
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
			"sensitive": sensitiveFunc(options.ShowSensitive),
		}).Parse(a)
		if err != nil {
			return "", nil, err
		}

		o := new(bytes.Buffer)

		if err := tpl.Execute(o, data); err != nil {
			return "", nil, err
		}

		args[i] = o.String()
	}

	return name, args, nil
}

// Display returns the command as a human readable string.
func (c Command) Display(data TemplateData) (string, error) {
	str := []string{}

	if c.Name != "" {
		str = append(str, chalk.Cyan.Color(c.Name))
	}

	name, args, err := c.Render(data, TemplateOptions{})
	if err != nil {
		return "", err
	}

	command := append([]string{name}, args...)
	str = append(str, chalk.Green.Color(strings.Join(command, " ")))

	return strings.Join(str, ": "), nil
}

// execute runs the command with the given config and outputs.
func (c Command) execute(ctx context.Context, config *Config, args Args, previousOutputs Outputs, streams *genericclioptions.IOStreams) (Outputs, error) {
	outputs := Outputs{}

	for k, v := range previousOutputs {
		outputs[k] = v
	}

	data := TemplateData{
		LocalPort: config.LocalPort,
		Args:      args,
		Outputs:   previousOutputs.Stdout(),
	}

	cmd, err := c.ToCmd(ctx, data)
	if err != nil {
		return nil, err
	}

	cmdStr, err := c.Display(data)
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(streams.ErrOut, "> %s\n", cmdStr)

	if c.Interactive {
		cmd.Stdout = streams.Out
		cmd.Stderr = streams.ErrOut
		cmd.Stdin = streams.In

		return outputs, cmd.Run()
	}

	outBuff := new(bytes.Buffer)
	errBuff := new(bytes.Buffer)

	ows := []io.Writer{outBuff}
	ews := []io.Writer{errBuff}

	if config.Verbose {
		ows = append(ows, streams.Out)
		ews = append(ews, streams.ErrOut)
	}

	cmd.Stdout = io.MultiWriter(ows...)
	cmd.Stderr = io.MultiWriter(ews...)

	if err := cmd.Run(); err != nil {
		name, args, _ := c.Render(data, TemplateOptions{
			ShowSensitive: false,
		})

		errStr := fmt.Sprintf("Error running command: %v\n%s\n", append([]string{name}, args...), errBuff)

		fmt.Fprint(streams.ErrOut, chalk.Red.Color(errStr))

		return nil, err
	}

	if c.ID != "" {
		outputs[c.ID] = Output{
			Stdout: outBuff.String(),
			Stderr: errBuff.String(),
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
