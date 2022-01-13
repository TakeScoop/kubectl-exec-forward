package command

import (
	"context"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Commands stores a slice of commands and provides some helper execution methods.
type Commands []*Command

// Execute runs each command in the calling slice sequentially using the passed config and the outputs accumulated to that point.
func (c Commands) Execute(ctx context.Context, config *Config, args Args, outputs Outputs, streams genericclioptions.IOStreams) (Outputs, error) {
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
