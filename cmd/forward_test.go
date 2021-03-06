package cmd

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/howeyc/fsnotify"
	"github.com/phayes/freeport"
	"github.com/stretchr/testify/assert"
	"github.com/takescoop/kubectl-exec-forward/internal/annotation"
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

type SafeBuffer struct {
	mutex sync.RWMutex
	buf   bytes.Buffer
}

func (b *SafeBuffer) Write(bs []byte) (int, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.buf.Write(bs)
}

func (b *SafeBuffer) String() string {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.buf.String()
}

func TestRunForwardCommand(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx := context.Background()

	clientset := getKubernetesClientset(t)

	ns, err := clientset.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{GenerateName: "test"}}, metav1.CreateOptions{})
	assert.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, clientset.CoreV1().Namespaces().Delete(ctx, ns.Name, metav1.DeleteOptions{}))
	})

	doneDir, err := os.MkdirTemp("", "test")
	assert.NoError(t, err)

	doneFile := fmt.Sprintf("%s/test", doneDir)

	pod, err := clientset.CoreV1().Pods(ns.Name).Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Annotations: map[string]string{
				annotation.PreConnect: `[{"command": ["echo", "test"]}]`,
				annotation.Command:    fmt.Sprintf(`{"command": ["touch", %q]}`, doneFile),
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Image: "nginx:1",
					Name:  "nginx",
					Ports: []corev1.ContainerPort{{
						ContainerPort: 80,
					}},
				},
			},
		},
	}, metav1.CreateOptions{})
	assert.NoError(t, err)

	waitForPod(ctx, t, clientset, pod)

	out := &SafeBuffer{}
	outErr := &SafeBuffer{}

	cmd := newForwardCommand(genericclioptions.IOStreams{
		Out:    out,
		ErrOut: outErr,
	}, "0.0.0")

	localPort := freeport.GetPort()

	cmd.SetArgs([]string{
		"--namespace",
		pod.Namespace,
		"--verbose",
		"--persist",
		fmt.Sprintf("pod/%s", pod.Name),
		fmt.Sprintf("%d:%d", localPort, pod.Spec.Containers[0].Ports[0].ContainerPort),
	})

	doneChan := make(chan bool)
	errChan := make(chan error)

	watcher, err := fsnotify.NewWatcher()
	assert.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, watcher.Close())
	})

	cancelCtx, cancel := context.WithCancel(ctx)

	t.Cleanup(func() {
		cancel()
	})

	go func() {
		if err := cmd.ExecuteContext(cancelCtx); err != nil {
			errChan <- err
		}
	}()

	go func() {
		if err := watcher.Watch(doneDir); err != nil {
			errChan <- err
		}
	}()

	go func() {
		err := waitForFile(watcher, doneFile, 10*time.Second)

		if err != nil {
			errChan <- err
		} else {
			doneChan <- true
		}
	}()

	waitForFinish(t, doneChan, errChan)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://localhost:%d", localPort), nil)
	assert.NoError(t, err)

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)

	assert.NoError(t, err)
	assert.NoError(t, resp.Body.Close())

	assert.Equal(t, 200, resp.StatusCode)

	assert.True(t, strings.HasPrefix(out.String(), "test"), "stdout did not contain hook command output")
	assert.True(t, strings.HasPrefix(outErr.String(), ""), "stderr was not empty")
}

func waitForFinish(t *testing.T, doneChan chan bool, errChan chan error) {
	t.Helper()

	for {
		select {
		case <-doneChan:
			return
		case err := <-errChan:
			assert.NoError(t, err)

			return
		}
	}
}

func waitForPod(ctx context.Context, t *testing.T, clientset *kubernetes.Clientset, pod *corev1.Pod) {
	t.Helper()

	assert.NoError(t, wait.PollImmediate(time.Second, time.Second*15, func() (bool, error) {
		pod, err := clientset.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		switch pod.Status.Phase {
		case corev1.PodRunning:
			return true, nil
		case corev1.PodPending, corev1.PodUnknown:
			return false, nil
		case corev1.PodFailed, corev1.PodSucceeded:
			return false, fmt.Errorf("pod not running, has status %s", pod.Status.Phase)
		}

		return false, nil
	}), "Timed out waiting for test pod to start")
}

func waitForFile(watcher *fsnotify.Watcher, fileName string, timeout time.Duration) error {
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case ev := <-watcher.Event:
			if ev.Name == fileName {
				return nil
			}
		case err := <-watcher.Error:
			return err
		case <-timer.C:
			return fmt.Errorf("timed out waiting for done file to be written: %s", fileName)
		}
	}
}

func getKubernetesClientset(t *testing.T) *kubernetes.Clientset {
	t.Helper()

	kc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	rc, err := kc.ClientConfig()
	assert.NoError(t, err)

	clientset, err := kubernetes.NewForConfig(rc)
	assert.NoError(t, err)

	return clientset
}
