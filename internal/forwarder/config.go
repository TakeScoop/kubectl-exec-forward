package forwarder

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
)

// Config contains the information required to satisfy a call to Forward.
type Config struct {
	Pod  *corev1.Pod
	Port string
}

// GetLocalPort returns the local ports from the Config port mapping.
func (c Config) GetLocalPort() (port int, err error) {
	localStr, _ := splitPort(c.Port)

	local, err := strconv.ParseInt(localStr, 10, 64)
	if err != nil {
		return 0, err
	}

	return int(local), nil
}

// NewConfig interacts with the Kubernetes API to find a pod and ports suitable for forwarding.
func (c Client) NewConfig(resource string, portMap string) (*Config, error) {
	namespace, _, err := c.userConfig.Namespace()
	if err != nil {
		return nil, err
	}

	obj, pod, err := c.AttachablePodForObjectFn(resource, namespace, c.timeout)
	if err != nil {
		return nil, err
	}

	port, err := c.translatePorts(obj, pod, portMap)
	if err != nil {
		return nil, err
	}

	return &Config{
		Pod:  pod,
		Port: port,
	}, nil
}
