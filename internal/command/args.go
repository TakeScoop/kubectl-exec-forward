package command

import (
	"encoding/json"
)

// Args store command specific arguments to be passed to the hook commands.
type Args map[string]string

// ParseArgsFromAnnotations parses key value pairs from the passed annotations map, adds any overrides passed and returns a new args map.
func ParseArgsFromAnnotations(annotations map[string]string, overrides map[string]string) (Args, error) {
	args := Args{}

	v, ok := annotations[ArgsAnnotation]
	if ok {
		if err := json.Unmarshal([]byte(v), &args); err != nil {
			return nil, err
		}
	}

	for k, v := range overrides {
		args[k] = v
	}

	return args, nil
}
