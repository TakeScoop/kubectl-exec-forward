package annotation

import (
	"encoding/json"

	"github.com/takescoop/kubectl-exec-forward/internal/command"
)

// ParseCommands returns a slice of commands parsed from an annotations map at the value "key".
func ParseCommands(annotations map[string]string, key string) (commands command.Commands, err error) {
	v, ok := annotations[key]
	if !ok {
		return commands, nil
	}

	if err := json.Unmarshal([]byte(v), &commands); err != nil {
		return nil, err
	}

	return commands, err
}
