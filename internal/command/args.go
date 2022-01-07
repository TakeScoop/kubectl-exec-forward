package command

import (
	"encoding/json"
)

// Args store command specific arguments to be passed to the hook commands.
type Args map[string]string

// ParseArgsFromAnnotations parses key value pairs from the passed annotations map, adds any overrides passed and returns a new args map.
func ParseArgsFromAnnotations(annotations map[string]string) (Args, error) {
	args := Args{}

	v, ok := annotations[ArgsAnnotation]
	if ok {
		if err := json.Unmarshal([]byte(v), &args); err != nil {
			return nil, err
		}
	}

	return args, nil
}

// Merge merges the provided overrides into the existing args, mutating the existing args.
func (a Args) Merge(overrides map[string]string) {
	for k, v := range overrides {
		a[k] = v
	}
}
