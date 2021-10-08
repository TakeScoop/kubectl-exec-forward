package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/phayes/freeport"
	"github.com/spf13/cobra"
	"github.com/takescoop/service-connect/pkg/forwarder"
	"github.com/takescoop/service-connect/pkg/open"
	"github.com/takescoop/service-connect/pkg/rds"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Spec struct {
	LocalHost string
	LocalPort int
}

func NewConnectCommand() *cobra.Command {
	overrides := clientcmd.ConfigOverrides{}

	cmd := &cobra.Command{
		Use:   "service-connect svc [flags]",
		Short: "Manage connections to different services",
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

			annoType, ok := service.Annotations["aws-con.service.kubernetes.io/type"]
			if !ok {
				return fmt.Errorf("aws-con.service.kubernetes.io/type not found")
			}

			annoMeta, ok := service.Annotations["aws-con.service.kubernetes.io/meta"]
			if !ok {
				return fmt.Errorf("aws-con.service.kubernetes.io/meta not found")
			}

			config, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return err
			}

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			handlers := forwarder.Handlers{
				OnReady: nil,
				OnStop:  func() { <-sigChan },
			}

			localPort, err := freeport.GetFreePort()
			if err != nil {
				return err
			}

			switch annoType {
			case "rds-iam":
				var meta *rds.Meta
				err := json.Unmarshal([]byte(annoMeta), &meta)
				if err != nil {
					return err
				}

				dbUser, _ := flags.GetString("db-user")

				client := rds.New(config.Region, config)

				dbAuth, err := client.GetDBCredentials(ctx, meta, dbUser)
				if err != nil {
					return err
				}

				dbURL := url.URL{
					Scheme: dbAuth.Scheme,
					Host:   fmt.Sprintf("localhost:%d", localPort),
					User:   url.UserPassword(dbUser, dbAuth.Password),
					Path:   dbAuth.DBName,
				}
				conn := dbURL.String()

				callOpen, _ := flags.GetBool("open")

				handlers.OnReady = func() {
					fmt.Println(conn)
					if callOpen {
						open.Open(conn)
					}
				}

				if err = forwarder.Forward(clientset, restConfig, service, localPort, dbAuth.Port, handlers); err != nil {
					return err
				}
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
