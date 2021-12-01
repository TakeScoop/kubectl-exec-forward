package command

import "encoding/json"

type Config struct {
	LocalPort int `json:"local_port"`
}

func ParseConfig(annotations map[string]string, overrides *Config) (config *Config, err error) {
	v, ok := annotations[configAnnotation]
	if !ok {
		return config, nil
	}

	if err := json.Unmarshal([]byte(v), &config); err != nil {
		return nil, err
	}

	if overrides != nil {
		config.addOverrides(overrides)
	}

	return config, nil
}

func (c *Config) addOverrides(override *Config) {
	if override.LocalPort > 0 {
		c.LocalPort = override.LocalPort
	}
}
