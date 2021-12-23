package command

// Hooks store information regarding command hooks.
type Hooks struct {
	Pre     Commands
	Post    Commands
	Command Command
}

const (
	preConnectHookType  string = "pre-connect"
	postConnectHookType string = "post-connect"
	commandHookType     string = "command"
)

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

	hooks := &Hooks{
		Pre:  pre,
		Post: post,
	}

	c, err := parseComand(annotations, CommandAnnotation)
	if err != nil {
		return nil, err
	}

	c.Interactive = true

	if config != nil {
		if len(config.Command) > 0 {
			c.Command = append(config.Command, c.Command[1:]...)
		}
	}

	hooks.Command = c

	return hooks, nil
}
