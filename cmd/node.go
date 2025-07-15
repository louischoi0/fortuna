package cmd

import (
	"github.com/spf13/cobra"
)


func CreateNodeStartCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "run",
                Short: "",
                RunE: func(cmd *cobra.Command, args []string) error {
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

