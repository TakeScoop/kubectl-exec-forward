package command

type Hooks struct {
	Pre  Commands
	Post Commands
}

// newHooks returns a new Hooks struct assembled from the passed annotations
func newHooks(annotations map[string]string) (*Hooks, error) {
	pre, err := parseCommands(annotations, preAnnotation)
	if err != nil {
		return nil, err
	}

	post, err := parseCommands(annotations, postAnnotation)
	if err != nil {
		return nil, err
	}

	return &Hooks{
		Pre:  pre,
		Post: post,
	}, nil
}
