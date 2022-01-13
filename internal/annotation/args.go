package annotation

import (
	"encoding/json"

	"github.com/takescoop/kubectl-exec-forward/internal/command"
)

// ParseArgs parses key value pairs from the passed annotations map, adds any overrides passed and returns a new args map.
func ParseArgs(annotations map[string]string) (command.Args, error) {
	args := command.Args{}

	v, ok := annotations[Args]
	if ok {
		if err := json.Unmarshal([]byte(v), &args); err != nil {
			return nil, err
		}
	}

	return args, nil
}
