package cmd

import (
	"fortuna/service"
	"fortuna/core/storage"
	"fortuna/rock"
	"log"
	"github.com/spf13/cobra"
)

func CreateReplicaRunCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "run replica",
		RunE: func(cmd *cobra.Command, args []string) error {
			masterNodeAddr, _ := cmd.Flags().GetString("maddr")
			data_dir, _ := cmd.Flags().GetString("dir")

			storage.SetRootDir(data_dir)
			rock.SetMetastoreDataRootDir(data_dir)

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
	cmd.Flags().StringP("dir", "d", "replica", "endpoint")
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


