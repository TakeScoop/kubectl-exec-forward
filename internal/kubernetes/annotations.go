package kubernetes

import (
	"context"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

// GetAnnotations queries for the passed resource and returns the annotations for the found object.
func (c Client) GetAnnotations(ctx context.Context, resource string) (map[string]string, error) {
	r := strings.Split(resource, "/")

	switch r[0] {
	case "svc":
		fallthrough
	case "service":
		{
			return c.getServiceAnnotations(ctx, r[1])
		}
	case "po":
		fallthrough
	case "pod":
		{
			return c.getPodAnnotations(ctx, r[1])
		}
	case "deploy":
		fallthrough
	case "deployment":
		{
			return c.getDeploymentAnnotations(ctx, r[1])
		}
	default:
		return nil, fmt.Errorf("resource type %s not supported", r[0])
	}
}

// getServiceAnnotations returns annotations from the passed resource.
func (c Client) getServiceAnnotations(ctx context.Context, name string) (map[string]string, error) {
	res := c.Opts.RESTClient.Get().
		Namespace(c.Opts.Namespace).
		Name(name).
		Resource("services").
		Do(ctx)

	if err := res.Error(); err != nil {
		return nil, err
	}

	svc := v1.Service{}

	if err := res.Into(&svc); err != nil {
		return nil, err
	}

	return svc.Annotations, nil
}

// getPodAnnotations returns annotations from the passed resource.
func (c Client) getPodAnnotations(ctx context.Context, name string) (map[string]string, error) {
	res := c.Opts.RESTClient.Get().
		Namespace(c.Opts.Namespace).
		Name(name).
		Resource("pods").
		Do(ctx)

	if err := res.Error(); err != nil {
		return nil, err
	}

	pod := v1.Pod{}

	if err := res.Into(&pod); err != nil {
		return nil, err
	}

	return pod.Annotations, nil
}

// getDeploymentAnnotations returns annotations from the passed resource.
func (c Client) getDeploymentAnnotations(ctx context.Context, name string) (map[string]string, error) {
	res := c.Opts.RESTClient.Get().
		Namespace(c.Opts.Namespace).
		Name(name).
		Resource("deployments").
		Do(ctx)

	if err := res.Error(); err != nil {
		return nil, err
	}

	d := appsv1.Deployment{}

	if err := res.Into(&d); err != nil {
		return nil, err
	}

	return d.Annotations, nil
}
