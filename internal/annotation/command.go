package annotation

import (
	"encoding/json"

	"github.com/takescoop/kubectl-exec-forward/internal/command"
)

// ParseCommand returns a Command from annotations storing a single command in json format.
func ParseCommand(annotations map[string]string) (command command.Command, err error) {
	v, ok := annotations[Command]
	if !ok {
		return command, nil
	}

	if err := json.Unmarshal([]byte(v), &command); err != nil {
		return command, err
	}

	return command, nil
}
