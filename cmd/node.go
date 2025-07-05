package cmd

import (
	"fortuna/core"

	"github.com/spf13/cobra"
)

func CreateNodeStartCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "run",
		Short: "start synapse",
		RunE: func(cmd *cobra.Command, args []string) error {
			n := core.NewBasicSynapse("test")
			n.Init("test")
			return nil
		},
	}

	cmd.Flags().StringP("port", "p", "4427", "pair")
	return cmd
}

func CreateNodeCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "node management cli",
	}

	cmd.AddCommand(CreateNodeStartCmd())

	return cmd
}
