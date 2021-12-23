package command

import (
	"context"
	"encoding/json"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Commands stores a slice of commands and provides some helper execution methods.
type Commands []*Command

type commandInputs []*commandInput

// execute runs each command in the calling slice sequentially using the passed config and the outputs accumulated to that point.
func (c Commands) execute(ctx context.Context, config *Config, args *Args, previousOutputs map[string]Output, streams *genericclioptions.IOStreams) (outputs map[string]Output, err error) {
	outputs = map[string]Output{}
	for k, v := range previousOutputs {
		outputs[k] = v
	}

	for _, command := range c {
		outputs, err = command.execute(ctx, config, args, outputs, streams)
		if err != nil {
			return nil, err
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

	var inputs commandInputs

	if err := json.Unmarshal([]byte(v), &inputs); err != nil {
		return nil, err
	}

	for _, c := range inputs {
		cmd := c.toCommand(getHookTypeFromAnnotationKey(key))
		commands = append(commands, &cmd)
	}

	return commands, err
}
