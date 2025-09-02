package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"fortuna/rpc"
	"fortuna/structure"
	"fortuna/swift"
	"fortuna/util"

	"github.com/spf13/cobra"
)

func CreateGenVectorAPI() *cobra.Command {
	var kernel_version string
	var endpoint string
	var size int64
	var seed string

	cmd := &cobra.Command{
		Use:   "genvector",
		Short: "generate vector",
		Args:  cobra.NoArgs,

		Run: func(cmd *cobra.Command, args []string) {

			omap := structure.NewOrderedMap()
			omap.Set("kernel_version", kernel_version)
			omap.Set("size", size)
			omap.Set("seed", seed)

			payload := omap.Ser()

			request := rpc.CreateRequest(swift.PacketTypeGenVectorRequest, endpoint, payload)
			response, err := request.Call()

			if err != nil {
				fmt.Println("error: ", err.Error())
			}

			if err == nil {
				var buf []byte

				json.Unmarshal(response.Payload, &buf)
				fmt.Println(string(response.Payload))
				fmt.Println(buf)
				fmt.Println(util.DecodeInt64Array(buf))
			}
		},
	}

	cmd.Flags().StringVarP(&kernel_version, "kernel_version", "k", "base:v.0.0", "kernel version")
	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", CLI_DEFAULT_ENDPOINT, "endpoint to connect")
	cmd.Flags().StringVarP(&seed, "seed", "s", "abc", "seed")
	cmd.Flags().Int64VarP(&size, "size", "z", 8, "size")

	return cmd
}

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

func CreateGenStateSeedAPI() *cobra.Command {
	var kernel_version string
	var payload string
	var endpoint string

	cmd := &cobra.Command{
		Use:   "seed",
		Short: "seed",
		Args:  cobra.NoArgs,

		Run: func(cmd *cobra.Command, args []string) {
			request := rpc.CreateRequest(swift.PacketTypeStateSeedAPIRequest, endpoint, payload)
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

	cmd.Flags().StringVarP(&kernel_version, "kernel_version", "k", "base:v.0.0", "kernel version")
	cmd.Flags().StringVarP(&payload, "payload", "p", "", "payload")
	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", CLI_DEFAULT_ENDPOINT, "endpoint to connect")

	return cmd
}

func CreateAPICMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api",
		Short: "api",
	}

	cmd.AddCommand(CreateGenStateSeedAPI())
	cmd.AddCommand(CreateGenVectorAPI())
	cmd.AddCommand(CreateUniverseInfoAPI())
	cmd.AddCommand(CreateListEventsAPI())

	return cmd
}
