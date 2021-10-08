package forwarder

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	helpers "k8s.io/kubectl/pkg/polymorphichelpers"
	"k8s.io/kubectl/pkg/util/podutils"
)

type Handlers struct {
	OnReady func()
	OnStop  func()
}

func dialer(k kubernetes.Clientset, config *rest.Config, namespace string, pod string) (httpstream.Dialer, error) {
	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return nil, err
	}

	return spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", k.RESTClient().Post().Prefix("api/v1").Resource("pods").Namespace(namespace).Name(pod).SubResource("portforward").URL()), nil
}

func Forward(k *kubernetes.Clientset, config *rest.Config, service *v1.Service, localPort int, handlers Handlers) error {

	namespace, selector, err := helpers.SelectorsForObject(service)
	if err != nil {
		return err
	}

	sortBy := func(pods []*v1.Pod) sort.Interface { return podutils.ByLogging(pods) }

	pod, _, err := helpers.GetFirstPod(k.CoreV1(), namespace, selector.String(), time.Duration(time.Second*30), sortBy)
	if err != nil {
		return err
	}

	dialer, err := dialer(*k, config, service.Namespace, pod.Name)
	if err != nil {
		return err
	}

	readyChan := make(chan struct{})
	go func() {
		<-readyChan
		handlers.OnReady()
	}()

	stopChan := make(chan struct{})
	go func() {
		handlers.OnStop()
		stopChan <- struct{}{}
	}()

	pf, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%s", localPort, &service.Spec.Ports[0].TargetPort)}, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}

	return pf.ForwardPorts()
}
