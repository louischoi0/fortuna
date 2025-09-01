package cmd

import (
	"fortuna/core/storage"
	"fortuna/rock"
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
			data_dir, _ := cmd.Flags().GetString("dir")

			storage.SetRootDir(data_dir)
			rock.SetMetastoreDataRootDir(data_dir)

			replica := service.NewReplica()
			replica.ActivateIndexer()
			replica.BootStrap()
			replica.LoadUniverse()

			err := replica.Connect(masterNodeAddr)
			if err != nil {
				log.Fatalf(err.Error())
			}

			err = replica.SyncAllSpaces(true)
			if err != nil {
				log.Fatalf(err.Error())
			}

			return replica.Run()
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
