package command

import (
	"encoding/json"
)

type Args map[string]interface{}

func ParseArgs(annotations map[string]string, overrides map[string]string) (args *Args, err error) {
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

func (a *Args) addOverrides(overrides map[string]string) {
	for k, v := range overrides {
		(*a)[k] = v
	}
}
