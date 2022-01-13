package command

import (
	"context"
	"encoding/json"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Commands stores a slice of commands and provides some helper execution methods.
type Commands []*Command

// Execute runs each command in the calling slice sequentially using the passed config and the outputs accumulated to that point.
func (c Commands) Execute(ctx context.Context, config *Config, args Args, outputs Outputs, streams *genericclioptions.IOStreams) (Outputs, error) {
	for _, command := range c {
		output, err := command.Execute(ctx, config, args, outputs, streams)
		if err != nil {
			return nil, err
		}

		if command.ID != "" {
			outputs = outputs.Append(command.ID, string(output))
		}
	}

	return outputs, nil
}

// parseCommands returns a slice of commands parsed from an annotations map at the value "key".
func parseCommands(annotations map[string]string, key string) (commands Commands, err error) {
	v, ok := annotations[key]
	if !ok {
		return commands, nil
	}

	if err := json.Unmarshal([]byte(v), &commands); err != nil {
		return nil, err
	}

	return commands, err
}
