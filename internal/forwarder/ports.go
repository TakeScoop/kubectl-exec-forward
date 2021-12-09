/*
Package forwarder contains a few functions that are borrowed from kubernetes/kubectl repo. They are private and cannot be used directly.
https://github.com/kubernetes/kubectl/blob/0b0920722212395d20fd0eecb5abf45ecb3e0cac/pkg/cmd/portforward/portforward.go
*/
package forwarder

import (
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/util"
)

// Translates the passed runtime object port mappings into ports that target the passed pod.
func (c Client) translatePorts(obj runtime.Object, pod *corev1.Pod, ports []string) ([]string, error) {
	switch t := obj.(type) {
	case *corev1.Service:
		return translateServicePortToTargetPort(ports, *t, *pod)
	default:
		return convertPodNamedPortToNumber(ports, *pod)
	}
}

// splitPort splits port string which is in form of [LOCAL PORT]:REMOTE PORT
// and returns local and remote ports separately.
func splitPort(port string) (local, remote string) {
	parts := strings.Split(port, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return parts[0], parts[0]
}

// Translates service port to target port
// It rewrites ports as needed if the Service port declares targetPort.
// It returns an error when a named targetPort can't find a match in the pod, or the Service did not declare
// the port.
func translateServicePortToTargetPort(ports []string, svc corev1.Service, pod corev1.Pod) ([]string, error) {
	var translated []string

	for _, port := range ports {
		localPort, remotePort := splitPort(port)

		// nolint:gosec
		portnum, err := strconv.Atoi(remotePort)
		if err != nil {
			svcPort, err := util.LookupServicePortNumberByName(svc, remotePort)
			if err != nil {
				return nil, err
			}

			portnum = int(svcPort)

			if localPort == remotePort {
				localPort = strconv.Itoa(portnum)
			}
		}

		containerPort, err := util.LookupContainerPortNumberByServicePort(svc, pod, int32(portnum))
		if err != nil {
			// can't resolve a named port, or Service did not declare this port, return an error
			return nil, err
		}

		// convert the resolved target port back to a string
		remotePort = strconv.Itoa(int(containerPort))

		if localPort != remotePort {
			translated = append(translated, fmt.Sprintf("%s:%s", localPort, remotePort))
		} else {
			translated = append(translated, remotePort)
		}
	}

	return translated, nil
}

// convertPodNamedPortToNumber converts named ports into port numbers
// It returns an error when a named port can't be found in the pod containers.
func convertPodNamedPortToNumber(ports []string, pod corev1.Pod) ([]string, error) {
	var converted []string

	for _, port := range ports {
		localPort, remotePort := splitPort(port)

		containerPortStr := remotePort

		_, err := strconv.Atoi(remotePort)
		if err != nil {
			containerPort, err := util.LookupContainerPortNumberByName(pod, remotePort)
			if err != nil {
				return nil, err
			}

			containerPortStr = strconv.Itoa(int(containerPort))
		}

		if localPort != remotePort {
			converted = append(converted, fmt.Sprintf("%s:%s", localPort, containerPortStr))
		} else {
			converted = append(converted, containerPortStr)
		}
	}

	return converted, nil
}
