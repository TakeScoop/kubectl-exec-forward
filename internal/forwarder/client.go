package forwarder

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/scheme"
	"k8s.io/kubectl/pkg/util"
	"k8s.io/kubectl/pkg/util/podutils"
)

type Client struct {
	clientset  *kubernetes.Clientset
	restConfig *rest.Config
	factory    cmdutil.Factory
	timeout    time.Duration
}

func NewClient(overrides clientcmd.ConfigOverrides, getter *cmdutil.MatchVersionFlags) (*Client, error) {
	kc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&overrides,
	)

	config, err := kc.ClientConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	factory := cmdutil.NewFactory(getter)

	return &Client{
		clientset:  clientset,
		restConfig: config,
		factory:    factory,
		timeout:    500,
	}, nil
}

func (c Client) GetResource(namespace string, resource string) (runtime.Object, error) {
	return c.factory.NewBuilder().
		WithScheme(scheme.Scheme, scheme.Scheme.PrioritizedVersionsAllGroups()...).
		ContinueOnError().
		NamespaceParam(namespace).
		DefaultNamespace().
		ResourceNames("pods", resource).
		Do().
		Object()
}

func (c Client) GetPod(obj runtime.Object) (*corev1.Pod, error) {
	switch t := obj.(type) {
	case *corev1.Pod:
		return t, nil
	}

	namespace, selector, err := polymorphichelpers.SelectorsForObject(obj)
	if err != nil {
		return nil, fmt.Errorf("cannot attach to %T: %v", obj, err)
	}

	sortBy := func(pods []*corev1.Pod) sort.Interface { return sort.Reverse(podutils.ActivePods(pods)) }
	pod, _, err := polymorphichelpers.GetFirstPod(c.clientset.CoreV1(), namespace, selector.String(), c.timeout, sortBy)
	return pod, err
}

func (c Client) TranslatePorts(obj runtime.Object, pod *corev1.Pod, ports []string) ([]string, error) {
	switch t := obj.(type) {
	case *corev1.Service:
		return translateServicePortToTargetPort(ports, *t, *pod)
	default:
		return convertPodNamedPortToNumber(ports, *pod)
	}
}

func (c Client) Forward(namespace string, podName string, ports []string, readyChan chan struct{}, stopChan chan struct{}, streams genericclioptions.IOStreams) error {
	transport, upgrader, err := spdy.RoundTripperFor(c.restConfig)
	if err != nil {
		return err
	}

	url := c.clientset.RESTClient().Post().Prefix("api/v1").Resource("pods").Namespace(namespace).Name(podName).SubResource("portforward").URL()

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)

	fw, err := portforward.New(dialer, ports, stopChan, readyChan, streams.Out, streams.ErrOut)
	if err != nil {
		return err
	}

	return fw.ForwardPorts()
}

func translateServicePortToTargetPort(ports []string, svc corev1.Service, pod corev1.Pod) ([]string, error) {
	var translated []string
	for _, port := range ports {
		localPort, remotePort := splitPort(port)

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

func splitPort(port string) (local, remote string) {
	parts := strings.Split(port, ":")
	if len(parts) == 2 {
		return parts[0], parts[1]
	}

	return parts[0], parts[0]
}
