package command

import (
	"encoding/json"
	"log"
)

type Commands []*Command

// Execute runs each command in the calling slice sequentially using the passed config and the outputs accumulated to that point
func (c Commands) Execute(config *Config, arguments *Args, outputs *Outputs) error {
	for _, command := range c {
		stdout, stderr, err := command.Execute(config, arguments, outputs)
		if err != nil {
			return err
		}

		// TODO: use a log level for this instead of a warning log
		if stderr.Len() > 0 {
			log.Printf("Warning: command %q wrote to untracked stderr:\n", command.Command)
			log.Println(stderr.String())
		}

		(*outputs)[command.ID] = stdout.String()
	}

	return nil
}

func ParseCommands(annotations map[string]string, key string) (commands Commands, err error) {
	v, ok := annotations[key]
	if !ok {
		return commands, nil
	}

	if err := json.Unmarshal([]byte(v), &commands); err != nil {
		return nil, err
	}

	return commands, err
}
