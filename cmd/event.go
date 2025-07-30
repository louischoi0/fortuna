package cmd

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/rpc"
	"fortuna/structure"

	"github.com/spf13/cobra"
)

func CreateEventEmitCMD() *cobra.Command {
	var interface_id string
	var space_id string
	var kernel_version string
	var endpoint string
	var topic string
	var subtopic string
	var tag string

	cmd := &cobra.Command{
		Use:   "emit",
		Short: "emit event",
		Args:  cobra.NoArgs,

		Run: func(cmd *cobra.Command, args []string) {
			params := structure.NewOrderedMap()
			params.Set("slot_count", 2)

			auth := ""
			payload := structure.NewOrderedMap()

			spec := model.NewEventSpec(interface_id, kernel_version, params)
			event := model.NewEventRequest(space_id, payload, spec, topic, subtopic, tag)

			request := rpc.CreateEventRequest(endpoint, event, auth)

			response, err := request.Call()

			if err != nil {
				fmt.Println("error: ", err.Error())
			}

			if err == nil {
				fmt.Println(string(response.Payload))
			}
		},
	}

	cmd.Flags().StringVarP(&endpoint, "space_id", "s", "", "space id")
	cmd.Flags().StringVarP(&interface_id, "interface_id", "i", "FIC_001", "interface id")
	cmd.Flags().StringVarP(&kernel_version, "kernel_version", "k", "base-v0.0.0", "interface id")

	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", CLI_DEFAULT_ENDPOINT, "endpoint to connect")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "topic")
	cmd.Flags().StringVarP(&subtopic, "subtopic", "c", "", "subtopic")
	cmd.Flags().StringVarP(&tag, "tag", "g", "", "tag")

	return cmd
}

func CreateEventCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "event",
	}

	cmd.AddCommand(CreateEventEmitCMD())

	return cmd
}
