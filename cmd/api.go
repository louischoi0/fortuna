package cmd

import (
	"github.com/spf13/cobra"
	"fortuna/rpc"
	"fortuna/swift"
	"fmt"
)

func CreateGenStateSeedAPI() *cobra.Command {
	var kernel_version	string
	var payload		string
	var endpoint 		string

	cmd := &cobra.Command{
		Use:   "seed",
		Short: "emit event",
		Args:  cobra.NoArgs,

		Run: func(cmd *cobra.Command, args []string) {
			request := rpc.CreateRequest(swift.PacketTypeStateSeedAPIRequest, endpoint, payload)
			response, err := request.Call()

                        if err != nil {
                                fmt.Println("error: ", err.Error())
                        }

                        if err == nil {
                                buf := swift.FormatResponse(&response.Payload)
                                fmt.Println(buf)
                        }

		},
        }

        cmd.Flags().StringVarP(&kernel_version, "kernel_version", "k", "base-v0.0.0", "kernel version")
        cmd.Flags().StringVarP(&payload, "payload", "k", "", "payload")
        cmd.Flags().StringVarP(&endpoint, "endpoint", "e", CLI_DEFAULT_ENDPOINT, "endpoint to connect")

        return cmd
}

func CreateAPICMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "api",
	}

	cmd.AddCommand(CreateGenStateSeedAPI())

	return cmd
}
