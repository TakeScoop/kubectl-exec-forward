package execforward

import (
	"github.com/takescoop/kubectl-exec-forward/internal/annotation"
	"github.com/takescoop/kubectl-exec-forward/internal/command"
)

// Hooks store information regarding command hooks.
type Hooks struct {
	Pre     command.Commands
	Post    command.Commands
	Command command.Command
}

// newHooks returns a new Hooks struct assembled from the passed annotations.
func newHooks(annotations map[string]string, config *Config) (*Hooks, error) {
	pre, err := annotation.ParseCommands(annotations, annotation.PreConnect)
	if err != nil {
		return nil, err
	}

	post, err := annotation.ParseCommands(annotations, annotation.PostConnect)
	if err != nil {
		return nil, err
	}

	hooks := &Hooks{
		Pre:  pre,
		Post: post,
	}

	c, err := annotation.ParseCommand(annotations)
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
