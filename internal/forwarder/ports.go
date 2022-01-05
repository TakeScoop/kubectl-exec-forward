/*
Package forwarder contains a few functions that are borrowed from kubernetes/kubectl repo. They are private and cannot be used directly.
https://github.com/kubernetes/kubectl/blob/0b0920722212395d20fd0eecb5abf45ecb3e0cac/pkg/cmd/portforward/portforward.go
*/
package forwarder

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/phayes/freeport"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/kubectl/pkg/util"
)

// Translates the passed runtime object port mappings into ports that target the passed pod.
func (c Client) translatePorts(obj runtime.Object, pod *corev1.Pod, port string) (string, error) {
	local, remote := splitPort(port)

	if local == "0" {
		newLocal, err := freeport.GetFreePort()
		if err != nil {
			return "", err
		}

		port = fmt.Sprintf("%d:%s", newLocal, remote)
	}

	switch t := obj.(type) {
	case *corev1.Service:
		return translateServicePortToTargetPort(port, *t, *pod)
	default:
		return convertPodNamedPortToNumber(port, *pod)
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
func translateServicePortToTargetPort(port string, svc corev1.Service, pod corev1.Pod) (string, error) {
	localPort, remotePort := splitPort(port)

	// nolint:gosec
	portnum, err := strconv.Atoi(remotePort)
	if err != nil {
		svcPort, err := util.LookupServicePortNumberByName(svc, remotePort)
		if err != nil {
			return "", err
		}

		portnum = int(svcPort)

		if localPort == remotePort {
			localPort = strconv.Itoa(portnum)
		}
	}

	containerPort, err := util.LookupContainerPortNumberByServicePort(svc, pod, int32(portnum))
	if err != nil {
		// can't resolve a named port, or Service did not declare this port, return an error
		return "", err
	}

	// convert the resolved target port back to a string
	remotePort = strconv.Itoa(int(containerPort))

	if localPort != remotePort {
		return fmt.Sprintf("%s:%s", localPort, remotePort), nil
	}

	return remotePort, nil
}

// convertPodNamedPortToNumber converts named ports into port numbers
// It returns an error when a named port can't be found in the pod containers.
func convertPodNamedPortToNumber(port string, pod corev1.Pod) (string, error) {
	localPort, remotePort := splitPort(port)

	containerPortStr := remotePort

	_, err := strconv.Atoi(remotePort)
	if err != nil {
		containerPort, err := util.LookupContainerPortNumberByName(pod, remotePort)
		if err != nil {
			return "", err
		}

		containerPortStr = strconv.Itoa(int(containerPort))
	}

	if localPort != remotePort {
		return fmt.Sprintf("%s:%s", localPort, containerPortStr), nil
	}

	return containerPortStr, nil
}
