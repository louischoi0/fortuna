package cmd

import (
	"fortuna/rpc"
	"fortuna/swift"
	"log"

	"github.com/spf13/cobra"
)

const CLI_DEFAULT_ENDPOINT = "localhost:4277"

func CreatePingRequestCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ping",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint, _ := cmd.Flags().GetString("endpoint")
			request := rpc.NewRawRequest(endpoint, swift.PacketTypePing, `""`)
			response, err := request.Call()
			if err != nil {
				log.Fatalf(err.Error())
			}

			buffer := swift.FormatJSONResponse(response.Payload)
			log.Println(buffer)

			return nil
		},
	}

	cmd.Flags().StringP("endpoint", "e", "0.0.0.0:4277", "endpoint")
	return cmd
}

func CreateNodeStartCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "run",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			/**
			workspace, _ := cmd.Flags().GetString("workspace")
			port, _ := cmd.Flags().GetInt("port")

			syn := component.NewSynapse(workspace)
			syn.StartUp()
			syn.Run(port)
			*/

			return nil
		},
	}

	cmd.Flags().StringP("workspace", "w", "abcd", "")
	cmd.Flags().IntP("port", "p", 4277, "")
	return cmd
}

func CreateNodeCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node",
		Short: "node management cli",
	}

	cmd.AddCommand(CreateNodeStartCmd())
	cmd.AddCommand(CreatePingRequestCMD())

	return cmd
}
