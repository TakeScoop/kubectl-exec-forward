package command

import (
	"context"
	"encoding/json"
)

type Commands []*Command

// Execute runs each command in the calling slice sequentially using the passed config and the outputs accumulated to that point
func (c Commands) Execute(ctx context.Context, config *Config, arguments *Args, outputs map[string]Output, ios IO) error {
	for _, command := range c {
		output, err := command.Execute(ctx, config, arguments, outputs, ios)
		if err != nil {
			return err
		}

		if command.ID != "" {
			outputs[command.ID] = output
		}
	}

	return nil
}

func ParseCommands(annotations map[string]string, key string) (commands Commands, err error) {
	v, ok := annotations[key]
	if !ok {
		return commands, nil
	}

	if err := json.Unmarshal([]byte(v), &commands); err != nil {
		return nil, err
	}

	return commands, err
}
