package command

// Hooks store information regarding command hooks.
type Hooks struct {
	Pre  Commands
	Post Commands
}

// newHooks returns a new Hooks struct assembled from the passed annotations.
func newHooks(annotations map[string]string) (*Hooks, error) {
	pre, err := parseCommands(annotations, PreAnnotation)
	if err != nil {
		return nil, err
	}

	post, err := parseCommands(annotations, PostAnnotation)
	if err != nil {
		return nil, err
	}

	return &Hooks{
		Pre:  pre,
		Post: post,
	}, nil
}
