package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/howeyc/fsnotify"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/takescoop/kubectl-port-forward-hooks/internal/command"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func TestParseArgs(t *testing.T) {
	t.Run("Parse basic args flag", func(t *testing.T) {
		args, err := parseArgs([]string{"foo=bar", "baz=woz"})
		assert.NoError(t, err)

		assert.Equal(t, map[string]string{
			"foo": "bar",
			"baz": "woz",
		}, args)
	})

	t.Run("Parse empty args flag", func(t *testing.T) {
		args, err := parseArgs([]string{})
		assert.NoError(t, err)

		assert.Equal(t, map[string]string{}, args)
	})

	t.Run("Error on malformed kv input", func(t *testing.T) {
		_, err := parseArgs([]string{"foo bar"})
		assert.Error(t, err)
	})
}

func waitForPod(ctx context.Context, clientset *kubernetes.Clientset, pod *corev1.Pod) error {
	return wait.PollImmediate(time.Second, time.Minute, func() (bool, error) {
		pod, err := clientset.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case corev1.PodRunning:
			log.Println("pod running, pod unknown")
			return true, nil
		case corev1.PodPending, corev1.PodUnknown:
			log.Println("pod pending, pod unknown")
			return false, nil
		case corev1.PodFailed, corev1.PodSucceeded:
			log.Println("pod failed, pod succeeded")
			return false, fmt.Errorf("pod not running, has status %s", pod.Status.Phase)
		}

		return false, nil
	})
}

func waitForFile(watcher *fsnotify.Watcher, fileName string, timeout time.Duration) error {
	fmt.Println("waiting for done file")
	for {
		timer := time.NewTimer(timeout)

		select {
		case ev := <-watcher.Event:
			fmt.Println("event:", ev)
			if ev.IsCreate() && ev.Name == fileName {
				return nil
			}
		case err := <-watcher.Error:
			return err
		case <-timer.C:
			return fmt.Errorf("timed out waiting for finish file to be written: %s", fileName)
		}

		timer.Stop()
	}
}

func TestForward(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	kc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	rc, err := kc.ClientConfig()
	assert.NoError(t, err)

	clientset, err := kubernetes.NewForConfig(rc)
	assert.NoError(t, err)

	timestamp := time.Now().Unix()

	ns, err := clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("test-%d", timestamp)}}, metav1.CreateOptions{})
	assert.NoError(t, err)

	defer func() {
		log.Println("CLEANUP")
		assert.NoError(t, clientset.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{}))
	}()

	doneFile := fmt.Sprintf("/tmp/test-%d", timestamp)

	pod, err := clientset.CoreV1().Pods(ns.Name).Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Annotations: map[string]string{
				command.PreAnnotation:  `[{"command": ["echo", "test"]}]`,
				command.PostAnnotation: fmt.Sprintf(`[{"command": ["touch", %q]}]`, doneFile),
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: "nginx",
					Name:  "nginx",
					Ports: []corev1.ContainerPort{{
						Name:          "http",
						ContainerPort: 80,
					}},
				},
			},
		},
	}, metav1.CreateOptions{})
	assert.NoError(t, err)

	log.Println("before wait")
	assert.NoError(t, waitForPod(ctx, clientset, pod))
	log.Println("after wait")

	out := new(bytes.Buffer)
	outErr := new(bytes.Buffer)

	cmd := newForwardCommand(&genericclioptions.IOStreams{
		Out:    out,
		ErrOut: outErr,
	})

	localPort := freeport.GetPort()

	cmd.SetArgs([]string{
		"--namespace",
		pod.Namespace,
		"--verbose",
		fmt.Sprintf("pod/%s", pod.Name),
		fmt.Sprintf("%d:%d", localPort, pod.Spec.Containers[0].Ports[0].ContainerPort),
	})

	doneChan := make(chan bool)
	errChan := make(chan error)

	watcher, err := fsnotify.NewWatcher()
	assert.NoError(t, err)

	cancelCtx, cancel := context.WithCancel(ctx)

	go func() {
		fmt.Println("starting cmd from goroutine")
		err := cmd.ExecuteContext(cancelCtx)
		fmt.Println("error running forward command")
		errChan <- err
	}()

	go func() {
		fmt.Println("starting watch on /tmp from goroutine")
		err := watcher.Watch("/tmp")
		fmt.Println("watcher error", err)
		errChan <- err
	}()

	go func() {
		fmt.Println("Starting wait for file go routine")
		err := waitForFile(watcher, doneFile, 10*time.Second)

		fmt.Println("wait for fileerror", err)

		if err != nil {
			log.Println("ERROR WAITING FOR FILE", err)
			errChan <- err
		} else {
			log.Println("DONE waiting for file")
			doneChan <- true
		}
	}()

waitForFinish:
	for {
		select {
		case <-doneChan:
			log.Println("done heard")
			break waitForFinish
		case err := <-errChan:
			log.Println("error heard")
			assert.NoError(t, err)

			break waitForFinish
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://localhost:%d", localPort), nil)
	assert.NoError(t, err)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)

	assert.NoError(t, err)
	assert.NoError(t, resp.Body.Close())

	assert.Equal(t, 200, resp.StatusCode)

	cancel()
	assert.NoError(t, watcher.Close())

	assert.True(t, strings.HasPrefix(out.String(), "test"))
	assert.True(t, strings.HasPrefix(outErr.String(), ""))
}
