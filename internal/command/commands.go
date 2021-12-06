package command

import (
	"context"
	"encoding/json"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

type Commands []*Command

// execute runs each command in the calling slice sequentially using the passed config and the outputs accumulated to that point.
func (c Commands) execute(ctx context.Context, config *Config, arguments *Args, outputs map[string]Output, streams *genericclioptions.IOStreams) error {
	for _, command := range c {
		output, err := command.execute(ctx, config, arguments, outputs, streams)
		if err != nil {
			return err
		}

		if command.ID != "" {
			outputs[command.ID] = output
		}
	}

	return nil
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