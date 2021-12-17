package command

import (
	"fmt"
	"strings"
)

// Hooks store information regarding command hooks.
type Hooks struct {
	Pre  Commands
	Post Commands
}

// newHooks returns a new Hooks struct assembled from the passed annotations.
func newHooks(annotations map[string]string, overrides *Hooks) (*Hooks, error) {
	pre, err := parseCommands(annotations, PreAnnotation)
	if err != nil {
		return nil, err
	}

	pre, err = addCommandOverrides(pre, overrides.Pre)
	if err != nil {
		return nil, err
	}

	post, err := parseCommands(annotations, PostAnnotation)
	if err != nil {
		return nil, err
	}

	post, err = addCommandOverrides(post, overrides.Post)
	if err != nil {
		return nil, err
	}

	return &Hooks{
		Pre:  pre,
		Post: post,
	}, nil
}

type prefixedCommand struct {
	position string
	id       string
	targetID string
	command  []string
}

func (pc prefixedCommand) toCommand() *Command {
	return &Command{
		ID:      pc.id,
		Command: pc.command,
	}
}

func newPrefixCommand(command *Command, parsed Commands) (prefixedCommand, error) {
	split := strings.Split(command.ID, ":")
	if len(split) < 2 || len(split) > 3 {
		return prefixedCommand{}, fmt.Errorf("prefixed commands must be in the format of position:targetID:[ID] (pre:foo:[id])")
	}

	position := split[0]
	target := split[1]

	if position == "" || target == "" {
		return prefixedCommand{}, fmt.Errorf("prefixed commands must supply non-empty positions and targets (position:targetID:[ID]): %s", command.ID)
	}

	var id string
	if len(split) == 3 {
		id = split[2]
	}

	found := false
	for _, p := range parsed {
		if target == p.ID {
			found = true
		}
	}
	if !found {
		return prefixedCommand{}, fmt.Errorf("prefixed commands must target a valid ID: %s", target)
	}

	return prefixedCommand{
		position: position,
		targetID: target,
		id:       id,
		command:  command.Command,
	}, nil
}

func overrideCommands(initial Commands, overrides Commands) (commands Commands) {
	for _, c := range initial {
		var found *Command
		for _, o := range overrides {
			if o.ID == c.ID {
				found = o
			}
		}
		if found != nil {
			commands = append(commands, found)
		} else {
			commands = append(commands, c)
		}
	}

	return commands
}

func addCommandOverrides(parsed Commands, overrides Commands) (Commands, error) {
	var commands Commands
	var toAdd Commands
	var prefixed Commands
	var toOverride Commands
	var prefixedCommands []prefixedCommand

	for _, o := range overrides {
		if o.ID == "" {
			toAdd = append(toAdd, o)
		} else if strings.HasPrefix(o.ID, "pre:") || strings.HasPrefix(o.ID, "post:") {
			prefixed = append(prefixed, o)
		} else {
			var found *Command
			for _, c := range parsed {
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

	assembled := append(overrideCommands(parsed, toOverride), toAdd...)

	for _, p := range prefixed {
		pc, err := newPrefixCommand(p, assembled)
		if err != nil {
			return nil, err
		}
		prefixedCommands = append(prefixedCommands, pc)
	}

	for _, c := range assembled {
		for _, p := range prefixedCommands {
			if p.targetID == c.ID && p.position == "pre" {
				commands = append(commands, p.toCommand())
			}
		}
		commands = append(commands, c)
		for _, p := range prefixedCommands {
			if p.targetID == c.ID && p.position == "post" {
				commands = append(commands, p.toCommand())
			}
		}
	}

	return commands, nil
}
