package command

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Commands stores a slice of commands and provides some helper execution methods.
type Commands []*Command

// execute runs each command in the calling slice sequentially using the passed config and the outputs accumulated to that point.
func (c Commands) execute(ctx context.Context, config *Config, args *Args, previousOutputs map[string]Output, streams *genericclioptions.IOStreams) (map[string]Output, error) {
	outputs := map[string]Output{}
	for k, v := range previousOutputs {
		outputs[k] = v
	}

	for _, command := range c {
		output, err := command.execute(ctx, config, args, outputs, streams)
		if err != nil {
			return nil, err
		}

		if command.ID != "" {
			outputs[command.ID] = output
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

// mergeCommands takes an initial list of commands and adds, or replaces items with commands from the passed.
func mergeCommands(initial Commands, overrides Commands) (Commands, error) {
	var toAdd Commands

	var toOverride Commands

	var prefixed Commands

	for _, o := range overrides {
		if o.ID == "" {
			toAdd = append(toAdd, o)
		} else if strings.HasPrefix(o.ID, "pre:") || strings.HasPrefix(o.ID, "post:") {
			prefixed = append(prefixed, o)
		} else {
			var found *Command

			for _, c := range initial {
				if c.ID == o.ID {
					found = o
				}
			}

			if found == nil {
				toAdd = append(toAdd, o)
			} else {
				toOverride = append(toOverride, o)
			}
		}
	}

	current := append(overrideCommands(initial, toOverride), toAdd...)

	if err := validatePrefixCommands(prefixed, current); err != nil {
		return nil, err
	}

	return mergePrefixedCommands(current, prefixed)
}

// overrideCommands replaces any matching overrides from the passed commands and returns the result.
func overrideCommands(current Commands, overrides Commands) Commands {
	commands := make(Commands, len(current))

	for i, c := range current {
		var found *Command

		for _, o := range overrides {
			if o.ID == c.ID {
				found = o
			}
		}

		if found != nil {
			commands[i] = found
		} else {
			commands[i] = c
		}
	}

	return commands
}

// mergePrefixedCommands merges the passed prefixed commands with the passed commands in the appropriate order and returns the result.
func mergePrefixedCommands(current Commands, prefixed Commands) (Commands, error) {
	commands := Commands{}

	for _, c := range current {
		for _, p := range prefixed {
			position, targetID, ID := parsePrefixedCommandID(p.ID)

			if targetID == c.ID && position == "pre" {
				commands = append(commands, &Command{ID: ID, Command: p.Command})
			}
		}

		commands = append(commands, c)

		for _, p := range prefixed {
			position, targetID, ID := parsePrefixedCommandID(p.ID)

			if targetID == c.ID && position == "post" {
				commands = append(commands, &Command{ID: ID, Command: p.Command})
			}
		}
	}

	return commands, nil
}

// validatePrefixCommands takes a list of commands with prefix IDs, ensures they are formatted correctly and that they have a matching current command to position against.
func validatePrefixCommands(prefixed Commands, current Commands) error {
	for _, p := range prefixed {
		split := strings.Split(p.ID, ":")
		if len(split) < 2 || len(split) > 3 {
			return fmt.Errorf("prefixed commands must be in the format of position:targetID:[ID] (pre:foo:[id])")
		}

		position := split[0]
		target := split[1]

		if position == "" || target == "" {
			return fmt.Errorf("prefixed commands must supply a non-empty position and target (position:targetID:[ID]): %s", p.ID)
		}

		found := false

		for _, c := range current {
			if target == c.ID {
				found = true
			}
		}

		if !found {
			return fmt.Errorf("prefixed commands must target a valid ID: %s", target)
		}
	}

	return nil
}

// parsePrefixedCommand takes a prefixed ID and returns the the parsed parts of the passed ID. It assumes that the ID is valid.
func parsePrefixedCommandID(cid string) (position string, targetID string, ID string) {
	split := strings.Split(cid, ":")

	if len(split) == 3 {
		ID = split[2]
	}

	return split[0], split[1], ID
}
