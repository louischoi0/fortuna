package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"fortuna/rpc"
	"fortuna/structure"
	"fortuna/swift"

	"github.com/spf13/cobra"
)

func CreateUniverseInfoAPI() *cobra.Command {
	var endpoint string

	cmd := &cobra.Command{
		Use:   "universe",
		Short: "get universe info",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			request := rpc.CreateRequest(swift.PacketTypeGETUniverseInfoRequest, endpoint, "")
			response, err := request.Call()

			if err != nil {
				fmt.Println("error: ", err.Error())
			}

			if err == nil {
				buf := swift.FormatJSONResponse(response.Payload)
				fmt.Println(buf)
			}
		},
	}

	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", CLI_DEFAULT_ENDPOINT, "endpoint to connect")

	return cmd
}

func CreateListEventsAPI() *cobra.Command {
	var spaceID  string
	var endpoint string

	cmd := &cobra.Command{
		Use:   "events",
		Short: "events",
		Args:  cobra.NoArgs,

		Run: func(cmd *cobra.Command, args []string) {
			payload := struct {
				SpaceID string 	`json:"space_id"`
				Count	int	`json:"count"`
			} {
				SpaceID: spaceID,
				Count: 100,
			}

			buf, err := json.Marshal(payload)
			if err != nil {
				log.Fatalf(err.Error())
			}

			request := rpc.CreateRequest(swift.PacketTypeListEventsRequest, endpoint, string(buf))
			response, err := request.Call()

			if err != nil {
				fmt.Println("error: ", err.Error())
			} else {
				buf := swift.FormatJSONResponse(response.Payload)
				fmt.Println(buf)
			}
		},
	}

	cmd.Flags().StringVarP(&spaceID, "space_id", "s", "0000000000000000000000000000000000000000000000000000000000000000", "spaceID")
	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", CLI_DEFAULT_ENDPOINT, "endpoint to connect")

	return cmd
}

func CreateVerifyEventAPI() *cobra.Command {
	var endpoint string
	var space_id string
	var event_result_hash string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "verify event result",
		Args:  cobra.NoArgs,

		Run: func(cmd *cobra.Command, args []string) {
			omap := structure.NewOrderedMap()
			omap.Set("space_id", space_id)
			omap.Set("event_hash", event_result_hash)

			payload := omap.Ser()

			request := rpc.CreateRequest(swift.PacketTypeVerifyEventResultRequest, endpoint, payload)
			response, err := request.Call()

			if err != nil {
				fmt.Println("error: ", err.Error())
			}

			if err == nil {
				buf := swift.FormatJSONResponse(response.Payload)
				fmt.Println(buf)
			}
		},
	}

	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", CLI_DEFAULT_ENDPOINT, "endpoint to connect")
	cmd.Flags().StringVarP(&space_id, "space_id", "s", "", "space_id")
	cmd.Flags().StringVarP(&event_result_hash, "event_result_hash", "a", "", "size")

	return cmd
}

func CreateAPICMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "api",
	}

	cmd.AddCommand(CreateUniverseInfoAPI())
	cmd.AddCommand(CreateListEventsAPI())
	cmd.AddCommand(CreateVerifyEventAPI())

	return cmd
}
