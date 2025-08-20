package cmd

import (
	"fortuna/service"
	"log"
	"github.com/spf13/cobra"
)

func CreateReplicaRunCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run replica",
		RunE: func(cmd *cobra.Command, args []string) error {
			masterNodeAddr, _ := cmd.Flags().GetString("maddr")
			replica := service.NewReplica()

			err := replica.Connect(masterNodeAddr)		
			if err != nil {
				log.Fatalf(err.Error())
			}

			replica.Run()
			return nil
		},
	}

	cmd.Flags().StringP("maddr", "a", CLI_DEFAULT_ENDPOINT, "endpoint")
	return cmd
}

func CreateReplicaCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replica",
		Short: "replica management cli",
	}

	cmd.AddCommand(CreateReplicaRunCMD())

	return cmd
}


