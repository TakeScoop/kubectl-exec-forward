package command

// Hooks store information regarding command hooks.
type Hooks struct {
	Pre     Commands
	Post    Commands
	Command Commands
}

// newHooks returns a new Hooks struct assembled from the passed annotations.
func newHooks(annotations map[string]string, config *Config) (*Hooks, error) {
	pre, err := parseCommands(annotations, PreAnnotation)
	if err != nil {
		return nil, err
	}

	post, err := parseCommands(annotations, PostAnnotation)
	if err != nil {
		return nil, err
	}

	command, err := parseComand(annotations, CommandAnnotation)
	if err != nil {
		return nil, err
	}

	hooks := &Hooks{
		Pre:  pre,
		Post: post,
	}

	if command != nil {
		if config != nil {
			command.Command = append(config.Command, command.Command[1:]...)
		}

		hooks.Command = Commands{command}
	}

	return hooks, nil
}
