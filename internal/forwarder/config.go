package forwarder

import (
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/kubectl/pkg/scheme"
)

// Config contains the information required to satisfy a call to Forward.
type Config struct {
	Pod   *corev1.Pod
	Ports []string
}

func (c Config) GetLocalPorts() (ports []int, err error) {
	for _, p := range c.Ports {
		localStr, _ := splitPort(p)

		local, err := strconv.ParseInt(localStr, 10, 64)
		if err != nil {
			return ports, err
		}

		ports = append(ports, int(local))
	}

	return ports, nil
}

// NewConfig interacts with the Kubernetes API to find a pod and ports suitable for forwarding.
func (c Client) NewConfig(resource string, portMap []string) (*Config, error) {
	namespace, _, err := c.userConfig.Namespace()
	if err != nil {
		return nil, err
	}

	obj, err := c.factory.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ContinueOnError().
		NamespaceParam(namespace).
		DefaultNamespace().
		ResourceNames("pods", resource).
		Do().
		Object()
	if err != nil {
		return nil, err
	}

	pod, err := c.getAttachablePod(obj)
	if err != nil {
		return nil, err
	}

	ports, err := c.translatePorts(obj, pod, portMap)
	if err != nil {
		return nil, err
	}

	return &Config{
		Pod:   pod,
		Ports: ports,
	}, nil
}
