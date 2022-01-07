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
	DisplayName string   `json:"name"`
}

type TemplateData struct {
	LocalPort int
	Args      Args
	Outputs   map[string]string
}

type TemplateOptions struct {
	ShowSensitive bool
}

// Name returns the name of the program.
func (c Command) Name() string {
	return c.Command[0]
}

// Args renders the command arguments using the provided template data and options.
func (c Command) Args(data TemplateData, options TemplateOptions) ([]string, error) {
	if len(c.Command) <= 1 {
		return []string{}, nil
	}

	args := c.Command[1:]

	for i, raw := range args {
		tpl, err := template.New(c.ID).Option("missingkey=error").Funcs(template.FuncMap{
			"trim":      strings.TrimSpace,
			"json":      gjson.Get,
			"sensitive": sensitiveFunc(options.ShowSensitive),
		}).Parse(raw)
		if err != nil {
			return nil, err
		}

		o := new(bytes.Buffer)

		if err := tpl.Execute(o, data); err != nil {
			return nil, err
		}

		args[i] = o.String()
	}

	return args, nil
}

// ToCmd returns a Cmd object that can be used with the exec package.
func (c Command) ToCmd(ctx context.Context, data TemplateData) (*exec.Cmd, error) {
	args, err := c.Args(data, TemplateOptions{
		ShowSensitive: true,
	})

	if err != nil {
		return nil, err
	}

	// nolint:gosec
	return exec.CommandContext(ctx, c.Name(), args...), nil
}

// Display returns the command as a human readable string.
func (c Command) Display(data TemplateData) (string, error) {
	str := []string{}

	if c.DisplayName != "" {
		str = append(str, chalk.Cyan.Color(c.DisplayName))
	}

	args, err := c.Args(data, TemplateOptions{})
	if err != nil {
		return "", err
	}

	command := append([]string{c.Name()}, args...)
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
		args, err := c.Args(data, TemplateOptions{
			ShowSensitive: false,
		})

		errStr := fmt.Sprintf("Error running command: %v\n%s\n", append([]string{c.Name()}, args...), errBuff)

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

// ParseCommandFromAnnotations returns a Command from annotations storing a single command in json format.
func ParseCommandFromAnnotations(annotations map[string]string) (command Command, err error) {
	v, ok := annotations[CommandAnnotation]
	if !ok {
		return command, nil
	}

	if err := json.Unmarshal([]byte(v), &command); err != nil {
		return command, err
	}

	return command, nil
}
