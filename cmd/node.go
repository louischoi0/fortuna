package cmd

import (
	"fortuna/core"
)


func CreateNodeStartCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "run",
                Short: "",
                RunE: func(cmd *cobra.Command, args []string) error {
			n := core.NewBasicNode()
			n.Init()
		}
	}

	cmd.Flags().StringP("port", "p", "4427", "pair")
}

func CreateNodeCMD() *cobra.Command {
        cmd := &cobra.Command{
                Use:   "node",
                Short: "node management cli",
        }

        cmd.AddCommand(CreateNodeStartCmd())

        return cmd
}

