package cmd

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/phayes/freeport"
	"github.com/takescoop/service-connect/pkg/forwarder"
	"github.com/takescoop/service-connect/pkg/open"
	"github.com/takescoop/service-connect/pkg/rds"

	"github.com/spf13/cobra"
)

func NewRDSCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rds",
		Short: "Manage an IAM authenticated RDS connection",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			flags := cmd.Flags()

			config, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return err
			}

			region, _ := flags.GetString("region")
			if region == "" {
				region = config.Region
			}

			app, _ := flags.GetString("app")
			environment, _ := flags.GetString("environment")
			user, _ := flags.GetString("user")
			callOpen, _ := flags.GetBool("open")

			client := rds.New(region, config)

			instance, err := client.GetDBInstanceByTags(ctx, []types.TagFilter{
				{
					Key:    aws.String("Environment"),
					Values: []string{environment},
				},
				{
					Key:    aws.String("App"),
					Values: []string{app},
				},
			})
			if err != nil {
				return err
			}

			token, err := client.GetAuthToken(ctx, instance, region, user, config)
			if err != nil {
				return err
			}

			localPort, err := freeport.GetFreePort()
			if err != nil {
				return err
			}

			conn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s", url.QueryEscape(user), url.QueryEscape(token), "localhost", localPort, *instance.DBName)

			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

			handlers := &forwarder.Handlers{
				OnReady: func() {
					fmt.Println(conn)
					if callOpen {
						open.Open(conn)
					}
				},
				OnStop: func() { <-sigChan },
			}

			if err = forwarder.Forward(ctx, *instance.Endpoint.Address, int(instance.Endpoint.Port), localPort, handlers); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String("region", "", "AWS region")
	cmd.Flags().String("user", "", "DB user")
	cmd.Flags().String("app", "", "Application associated with the database, identified by the AWS resource tag App=<value>")
	cmd.Flags().String("environment", "", "Environment associated with the database, identified by the AWS resource tag Environment=<value>")
	cmd.Flags().Bool("open", false, "Whether to call the linux open on the database connection string")

	return cmd
}
