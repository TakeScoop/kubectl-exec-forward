# kubectl-exec-forward

A kubectl plugin to run exec commands exposed on Kubernetes pod annotations during a port-forward action lifecycle. 

## Install

```sh
brew install takescoop/formulas/kubectl-exec-forward
```

## Usage

```sh
kubectl exec-forward type/name port [flags] [-- command]
```

### Flags

In addition to the standard kubectl flags, the following flags are available for the plugin

| Flag | Short | Description | Default |
|---|---|---|---|
| `--arg` | `-a` | `key=value` arguments passed to commands | `[]` |
| `--verbose` |`-v`| Whether to log verbosely |`false` |
| `--pod-timeout` | `-t` | Time to wait for an attachable pod to become available | `500` (ms) |
| `--persist` | `-p` | Whether to persist the forwarding connection after the main command has finished | `false` |

## Administration

Administrators can store complex behavior in Kubernetes pod annotations, allowing users to run a single `kubectl` command to interact with remote resources.

### Lifecycle

| Name | Description | 
|---|---|
| `pre-connect` | Run before establishing a port-forwarding connection |
| `post-connect` | Run after establishing a port-forwarding connection |
| `command` | The main command, run after `post-connect`. When `command` finishes, the port-forwarding connection is closed | 

### Annotations

| Name | Description | 
|---|---|
| `exec-forward.pod.kubernetes.io/args` | Arguments passed to commands for rendering, can be overridden from the CLI via `--arg\|-a` |
| `exec-forward.pod.kubernetes.io/pre-connect` | A JSON formatted list of commands executed before establishing a port-forwarding connection |
| `exec-forward.pod.kubernetes.io/post-connect` | A JSON formatted list of commands executed after establishing a port-forwarding connection |
| `exec-forward.pod.kubernetes.io/command` | A single JSON formatted command ran after `post-connect` |

#### Command

##### Object

The command object represents an exec command and associated configuration to be executed by the plugin during the port-forwarding lifecycle.

| Attribute | Description | Optional |
| --- | --- | --- |
| `id` | A unique identifier that can be used in subsequent commands to reference a previous command's output | `true` |
| `command` | The command to run as an array of strings | `false` |
| `interactive` | Whether the command should be run in interactive mode and can receive user input. Default is `false`. Note: the main `command` is always run in interactive mode | `true` |
| `name` | The display name for the command, shown during execution | `false` |

##### Rendering

Exec commands are rendered via [Go templates](https://pkg.go.dev/text/template) using inputs from the CLI, the `args` annotation, configuration from the port-forwarding connection and previous command outputs.

| Namespace | Description | Example |
|---|---|---|
| `.Args` | Arguments read from the `args` annotation and overridden using the `--arg\|-a` CLI flags | `{{.Args.username}}` |
| `.Outputs` | Stdout and Stderr from previously ran commands, stored by command `id` | `{{.Outputs.foo.Stdout}}` |
| `.Config` | Port-forwarding configuration | `{{.Config.LocalPort }}` |

##### Template functions

In addition to the following, all standard Go templating functions are available for use

| Name | Description |
| --- | --- |
| `sensitive` | Replaces the passed value with `********` when printing to console |
| `trim` | Removes white space from the beginning and end of a string, useful when piping to other template functions |

## Examples

Connect to an AWS RDS database using a generated password

```yaml
kind: Pod
metadata:
  name: db
  annotations:
    exec-forward.pod.kubernetes.io/args: '{"username":"read"}'
    exec-forward.pod.kubernetes.io/pre-connect: |
      [{
        "command": [
          "aws",
          "rds",
          "generate-db-auth-token",
          "--host",
          "...rds.amazonaws.com",
          "--port",
          "5432",
          "--username",
          "{{.Args.username}}"
        ],
        "id": "password",
        "name": "Generate temporary password"
      }]'
    exec-forward.pod.kubernetes.io/command: |
      {
        "command":[
          "psql",
          "postgres://{{.Args.username}}:{{ trim .Outputs.password.Stdout | urlquery | sensitive }}@localhost:{{.Config.LocalPort}}/db"
        ]
      }
spec:
  containers:
    image: alpine/socat
    - command:
      - socat
      - tcp-listen:5432,fork,reuseaddr
      - tcp-connect:...rds.amazonaws.com:5432
      ports:
        name: postgres
```

```sh
kubectl exec-forward pod/db postgres
```

Request data through a forwarded connection using a token generated

```yaml
kind: Deployment
metadata:
  name: foo
...
spec:
  template:
    metadata:
      annotations:
        exec-forward.pod.kubernetes.io/pre-connect: |
          [{
            "command": [
              "curl",
              "-d",
              "user=user&pass={{.Args.password}}",
              "https://example.com/token"
            ],
            "name": "Generate token",
            "id": token"
          }]
        exec-forward.pod.kubernetes.io/command: |
          {
            "command": [
              "curl",
              "-H",
              "X-Auth-Token: {{.Outputs.token.Stdout}}",
              "localhost:{{.Config.LocalPort}}/data"
            ]
          }
      spec:
        containers:
          ...
          - ports:
              name: http
              ...
```

```sh
kubectl port-forward deployment/foo http
```
