package command

import (
	"fmt"
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
