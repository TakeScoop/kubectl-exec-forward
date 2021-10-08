module github.com/takescoop/service-connect

go 1.16

require (
	github.com/aws/aws-sdk-go-v2 v1.9.1
	github.com/aws/aws-sdk-go-v2/config v1.8.2
	github.com/aws/aws-sdk-go-v2/feature/rds/auth v1.1.7
	github.com/aws/aws-sdk-go-v2/service/rds v1.9.0
	github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi v1.5.1
	github.com/bendrucker/kubernetes-port-forward-remote v0.0.0-20211004230343-e78275f0bb8b
	github.com/phayes/freeport v0.0.0-20180830031419-95f893ade6f2
	github.com/spf13/cobra v1.1.3
	github.com/traefik/yaegi v0.10.0
	k8s.io/api v0.22.2 // indirect
	k8s.io/apimachinery v0.22.2 // indirect
	k8s.io/cli-runtime v0.22.2
	k8s.io/client-go v0.22.2
	k8s.io/kubectl v0.22.2 // indirect
)

replace github.com/bendrucker/kubernetes-port-forward-remote v0.0.0-20211004230343-e78275f0bb8b => github.com/ryanwholey/kubernetes-port-forward-remote v0.0.0-20211006041054-a752512de576
