package kubernetes

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetPodAnnotations finds an attachable pod from the passed type/name, and returns the annotations from that pod.
func (c Client) GetPodAnnotations(ctx context.Context, resource string) (map[string]string, error) {
	p, err := c.Opts.PodClient.Pods(c.Opts.Namespace).Get(ctx, c.Opts.PodName, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return p.Annotations, nil
}
