package cmd

import (
	"github.com/spf13/cobra"
	"fortuna/core"
	"fortuna/swift"
	"fortuna/rpc"
	"fmt"
	"log"
)

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

			fmt.Printf(string(response.Payload))
			// swift.FormatResponse(&response.Payload)

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
			workspace, _ := cmd.Flags().GetString("workspace")
			port, _ := cmd.Flags().GetInt("port")

			syn := core.NewBasicSynapse(workspace)
			syn.StartUp()
			syn.Run(port)
	
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
