package cmd

import (
	"fortuna/rpc"
	"fortuna/service"
	"fortuna/swift"
	"log"

	"github.com/spf13/cobra"
)

const CLI_DEFAULT_ENDPOINT = "localhost:4277"

func CreateOracleListSpacesCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spaces",
		Short: "list all spaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint, _ := cmd.Flags().GetString("endpoint")
			request := rpc.NewRawRequest(endpoint, swift.PacketTypeOracleGETSpaceRequest, `""`)
			response, err := request.Call()
			if err != nil {
				log.Fatalf(err.Error())
			}

			buffer := swift.FormatJSONResponse(response.Payload)
			log.Println(buffer)

			return nil
		},
	}

	cmd.Flags().StringP("endpoint", "e", CLI_DEFAULT_ENDPOINT, "endpoint")
	return cmd
}

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

	cmd.Flags().StringP("endpoint", "e", CLI_DEFAULT_ENDPOINT, "endpoint")
	return cmd
}

func CreateOracleStartCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "run",
		Short: "",
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceID, _ := cmd.Flags().GetString("space_id")
			port, _ := cmd.Flags().GetInt("port")

			oracle := service.GetOracleService(spaceID, service.OracleConfig{
				MaxEventRequestsPerMinute: 1000,
				EventBufferSize:           1000,
				TransactionBufferSize:     1000,
			})

			if err := oracle.Bootstrap(); err != nil {
				log.Fatalf("Failed to bootstrap: %v", err.Error())
			}

			if err := oracle.Run(port); err != nil {
				log.Fatalf("Failed to run: %v", err.Error())
			}

			return nil
		},
	}

	cmd.Flags().StringP("space_id", "s", DEFAULT_SPACE_ID, "space id")
	cmd.Flags().IntP("port", "p", 4277, "port")
	return cmd
}

func CreateOracleCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle",
		Short: "oracle management cli",
	}

	cmd.AddCommand(CreateOracleStartCmd())
	cmd.AddCommand(CreatePingRequestCMD())
	cmd.AddCommand(CreateOracleListSpacesCMD())

	return cmd
}
