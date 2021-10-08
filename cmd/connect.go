package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/phayes/freeport"
	"github.com/spf13/cobra"
	"github.com/takescoop/service-connect/pkg/command"
	"github.com/takescoop/service-connect/pkg/forwarder"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Spec struct {
	LocalHost string
	LocalPort int
}

var annotation string = "aws-con.service.kubernetes.io"

func NewConnectCommand() *cobra.Command {
	overrides := clientcmd.ConfigOverrides{}

	cmd := &cobra.Command{
		Use:   "kubectl aws-connect svc [flags]",
		Short: "Manage connections to different AWS services",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			flags := cmd.Flags()

			kc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				clientcmd.NewDefaultClientConfigLoadingRules(),
				&overrides,
			)

			restConfig, err := kc.ClientConfig()
			if err != nil {
				return err
			}

			clientset, err := kubernetes.NewForConfig(restConfig)
			if err != nil {
				return err
			}

			namespace, _, err := kc.Namespace()
			if err != nil {
				return err
			}

			service, err := clientset.CoreV1().Services(namespace).Get(ctx, args[0], v1.GetOptions{})
			if err != nil {
				return err
			}

			var preCommands []command.Command

			annoPreCommands, ok := service.Annotations[fmt.Sprintf("%s/pre-commands", annotation)]
			if !ok {
				preCommands = []command.Command{}
			} else {
				if err := json.Unmarshal([]byte(annoPreCommands), &preCommands); err != nil {
					return err
				}
			}

			var options map[string]interface{}

			annoDefaults, ok := service.Annotations[fmt.Sprintf("%s/defaults", annotation)]
			if !ok {
				options = map[string]interface{}{}
			} else {
				if err := json.Unmarshal([]byte(annoDefaults), &options); err != nil {
					return err
				}
			}

			options["Pre"] = map[string]interface{}{}

			var postCommands []command.Command

			annoPostCommands, ok := service.Annotations[fmt.Sprintf("%s/post-commands", annotation)]
			if !ok {
				postCommands = []command.Command{}
			} else {
				if err := json.Unmarshal([]byte(annoPostCommands), &postCommands); err != nil {
					return err
				}
			}

			localPort, err := freeport.GetFreePort()
			if err != nil {
				return err
			}

			options["localPort"] = localPort

			user, _ := flags.GetString("db-user")
			if user != "" {
				options["username"] = user
			}

			for _, c := range preCommands {
				stdout, stderr, err := c.Execute(options)
				if err != nil {
					return err
				}

				if stderr.Len() > 0 {
					fmt.Println(stderr.String())
					return fmt.Errorf("failed to execute command %s:%s", c.ID, c.Command)
				}

				options["Pre"].(map[string]interface{})[c.ID] = stdout.String()
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			errChan := make(chan error)

			handlers := forwarder.Handlers{
				OnReady: func() {
					for _, c := range postCommands {
						stdout, stderr, err := c.Execute(options)
						if err != nil {
							errChan <- err
						}

						if stderr.Len() > 0 {
							fmt.Println(stderr.String())
							errChan <- fmt.Errorf("failed to execute command %s:%s", c.ID, c.Command)
						}

						fmt.Println(stdout.String())
					}
				},
				OnStop: func() { <-sigChan },
			}

			if err = forwarder.Forward(clientset, restConfig, service, localPort, handlers); err != nil {
				return err
			}

			return nil
		},
	}

	clientcmd.BindOverrideFlags(&overrides, cmd.PersistentFlags(), clientcmd.RecommendedConfigOverrideFlags(""))
	cmd.Flags().String("db-user", "", "DB user")
	cmd.Flags().Bool("open", false, "Whether to call the linux open on the database connection string")

	return cmd
}

func Execute() {
	cmd := NewConnectCommand()

	cobra.CheckErr(cmd.Execute())
}
