package command

import (
	"encoding/json"
)

// Args store command specific arguments to be passed to the hook commands.
type Args map[string]string

// parseArgs parses key value pairs from the passed annotations map, adds any overrides passed and returns a new args map.
func parseArgs(annotations map[string]string, overrides map[string]string) (args *Args, err error) {
	v, ok := annotations[argsAnnotation]
	if !ok {
		return args, err
	}

	if err := json.Unmarshal([]byte(v), &args); err != nil {
		return nil, err
	}

	args.addOverrides(overrides)

	return args, err
}

// addOverrides adds the passed key value pairs to the calling args map.
func (a *Args) addOverrides(overrides map[string]string) {
	for k, v := range overrides {
		(*a)[k] = v
	}
}
