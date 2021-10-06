package cmd

import "github.com/spf13/cobra"

func NewConnectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "service-connect",
		Short: "Manage connections to different services",
	}
}

func Execute() {
	cmd := NewConnectCommand()

	cmd.AddCommand(NewRDSCommand())

	cobra.CheckErr(cmd.Execute())
}
