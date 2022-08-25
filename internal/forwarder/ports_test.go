package forwarder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func TestTranslateServicePortToTargetPort(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string

		port    string
		service corev1.Service
		pod     corev1.Pod

		expected string
		error    string
	}{
		{
			name: "basic",
			port: "3000",
			service: corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port:       3000,
							TargetPort: intstr.FromInt(4000),
						},
					},
				},
			},
			expected: "3000:4000",
		},
		{
			name: "same local and remote",
			port: "3000",
			service: corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port:       3000,
							TargetPort: intstr.FromInt(3000),
						},
					},
				},
			},
			expected: "3000",
		},
		{
			name: "number to name",
			port: "3000",
			service: corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port:       3000,
							TargetPort: intstr.FromString("http"),
						},
					},
				},
			},
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 4000,
								},
							},
						},
					},
				},
			},
			expected: "3000:4000",
		},
		{
			name: "name to number",
			port: "http",
			service: corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:       "http",
							Port:       3000,
							TargetPort: intstr.FromInt(4000),
						},
					},
				},
			},
			expected: "3000:4000",
		},
		{
			name: "name not found",
			port: "http",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-svc",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Name:       "grpc",
							Port:       3000,
							TargetPort: intstr.FromInt(4000),
						},
					},
				},
			},
			error: `Service 'my-svc' does not have a named port 'http'`,
		},
		{
			name: "number not found",
			port: "3000",
			service: corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-svc",
				},
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{
						{
							Port:       3001,
							TargetPort: intstr.FromInt(4000),
						},
					},
				},
			},
			error: "Service my-svc does not have a service port 3000",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual, err := translateServicePortToTargetPort(tc.port, tc.service, tc.pod)

			if tc.error != "" {
				assert.EqualError(t, err, tc.error)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
