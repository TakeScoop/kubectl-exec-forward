package command

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strings"
	"text/template"
)

type Command struct {
	ID      string `json:"id"`
	Command string `json:"command"`
}

func (c Command) template(options map[string]interface{}) (string, error) {
	tpl := template.Must(template.New(c.ID).Funcs(template.FuncMap{
		"trim": strings.TrimSpace,
	}).Parse(string(c.Command)))

	out := new(bytes.Buffer)

	if err := tpl.Execute(out, options); err != nil {
		return "", err
	}

	return out.String(), nil
}

func (c Command) Execute(options map[string]interface{}) (stdout *bytes.Buffer, stderr *bytes.Buffer, err error) {
	command, err := c.template(options)
	if err != nil {
		return nil, nil, err
	}

	var parsed []string

	if err := json.Unmarshal([]byte(command), &parsed); err != nil {
		return nil, nil, err
	}

	name, args := parsed[0], parsed[1:]

	cmd := exec.Command(name, args...)

	stdout = new(bytes.Buffer)
	stderr = new(bytes.Buffer)

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	if err = cmd.Run(); err != nil {
		return nil, nil, err
	}

	return stdout, stderr, nil
}
