package cmd

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/core/vm"
	"fortuna/rpc"
	"fortuna/structure"
	"log"

	"github.com/spf13/cobra"
)

const DEFAULT_SPACE_ID = "0000000000000000000000000000000000000000000000000000000000000000"
const DEFAULT_PUBLISHER = "0000000000000000000000000000000000000000000000000000000000000000"

func CreateEventVerifyCMD() *cobra.Command {
	var event_payload string
	var event_hash string

	cmd := &cobra.Command{
		Use:   "verify",
		Short: "verify event",
		Args:  cobra.NoArgs,

		Run: func(cmd *cobra.Command, args []string) {
			omap, err := structure.ParseOrderedMap(string(event_payload))

			if err != nil {
				log.Fatalf(err.Error())
			}

			event, err := model.NewEventFromOrderedMap(omap)

			if err != nil {
				log.Fatalf(err.Error())
			}
			res := event.Verify(event_hash)
			if res {
				fmt.Println("event hash matched and verified")
			} else {
				fmt.Println("invalid hash")
			}
		},
	}

	cmd.Flags().StringVarP(&event_payload, "event_payload", "l", "", "payload")
	cmd.Flags().StringVarP(&event_hash, "event_hash", "a", "", "interface id")

	cmd.MarkFlagRequired("event_payload")
	cmd.MarkFlagRequired("event_hash")

	return cmd
}

func CreateEventEmitCMD() *cobra.Command {
	var interface_id string
	var space_id string
	var kernel_version string
	var endpoint string
	var topic string
	var subtopic string
	var tag string
	var publisher string

	cmd := &cobra.Command{
		Use:   "emit",
		Short: "emit event",
		Args:  cobra.NoArgs,

		Run: func(cmd *cobra.Command, args []string) {
			params := structure.NewOrderedMap()
			params.Set("slot_count", 64)

			auth := ""
			payload := structure.NewOrderedMap()
			payload.Set("a", 3)

			spec := model.NewEventSpec(interface_id, kernel_version, params)
			event, err := model.NewEventRequest(publisher, space_id, payload, spec, topic, subtopic, tag)

			if err != nil {
				log.Fatalf(err.Error())
			}

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

	cmd.Flags().StringVarP(&space_id, "space_id", "s", DEFAULT_SPACE_ID, "space id")
	cmd.Flags().StringVarP(&interface_id, "interface_id", "i", string(vm.FIC__001), "interface id")
	cmd.Flags().StringVarP(&kernel_version, "kernel_version", "k", string(vm.BaseV000), "kernel version")
	cmd.Flags().StringVarP(&endpoint, "endpoint", "e", CLI_DEFAULT_ENDPOINT, "endpoint to connect")
	cmd.Flags().StringVarP(&topic, "topic", "t", "", "topic")
	cmd.Flags().StringVarP(&subtopic, "subtopic", "c", "", "subtopic")
	cmd.Flags().StringVarP(&tag, "tag", "g", "", "tag")
	cmd.Flags().StringVarP(&publisher, "publisher", "p", DEFAULT_PUBLISHER, "publisher")

	return cmd
}

func CreateEventCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "event",
		Short: "event",
	}

	cmd.AddCommand(CreateEventEmitCMD())
	cmd.AddCommand(CreateEventVerifyCMD())

	return cmd
}
