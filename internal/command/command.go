package command

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
)

type Config map[string]interface{}

// GetInt returns the value at the passed key as an int, 0 if its not found
func (c Config) GetInt(key string) (int, error) {
	val, ok := c[key]
	if !ok {
		return 0, nil
	}

	if num, ok := val.(float64); ok {
		return int(num), nil

	}

	if numStr, ok := val.(string); ok {
		num64, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return 0, err
		}
		return int(num64), nil
	}

	return 0, fmt.Errorf("failed to convert %s to int: %v", key, val)
}

type Command struct {
	ID      string   `json:"id"`
	Command []string `json:"command"`
}

type Outputs map[string]string

type Commands []*Command

// Execute runs each command in the calling slice sequentially using the passed config and the outputs accumulated to that point
func (c Commands) Execute(config *Config, outputs *Outputs) error {
	for _, command := range c {
		stdout, stderr, err := command.Execute(config, outputs)
		if err != nil {
			return err
		}

		// TODO: use a log level for this instead of a warning log
		if stderr.Len() > 0 {
			log.Printf("Warning: command %q wrote to untracked stderr:\n", command.Command)
			log.Println(stderr.String())
		}

		(*outputs)[command.ID] = stdout.String()

		fmt.Println((*outputs)[command.ID])
	}

	return nil
}

// Execute runs the command with the given config and outputs
func (c Command) Execute(opts *Config, outputs *Outputs) (stdout *bytes.Buffer, stderr *bytes.Buffer, err error) {
	// TODO: run the command, return the output
	return &bytes.Buffer{}, &bytes.Buffer{}, nil
}
