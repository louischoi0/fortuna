package cmd

import (
	"github.com/spf13/cobra"
	"fortuna/structure"
	"fortuna/core/model"
	// "fortuna/swift"
	"log"
	"fortuna/rpc"
)

const CLI_DEFAULT_ENDPOINT = "localhost:4277"

func CreateEventEmitCMD() *cobra.Command {
        var interface_id 	string
	var space_id		string
	var kernel_version	string
	var endpoint 		string
	var topic		string
	var subtopic		string
	var tag 		string

	cmd := &cobra.Command{
		Use:   "emit",
		Short: "emit event",
		Args:  cobra.NoArgs,

		Run: func(cmd *cobra.Command, args []string) {
			params := structure.NewOrderedMap()
			auth := ""
			payload := structure.NewOrderedMap()

			spec := model.NewEventSpec(interface_id, kernel_version, params)
			event := model.NewEventRequest(space_id, payload, spec, topic, subtopic, tag)

			request := rpc.CreateEventRequest(endpoint, event, auth)

			log.Println(request.Payload)
			_, err := request.Call()

			if err != nil {
				log.Println("error: ", err.Error())
			}
			/**
			if err != nil {
				swift.FormatResponse(&response.Payload)
			} else {
				log.Printf(err.Error())
			}
			**/
		},
        }

        cmd.Flags().StringVarP(&endpoint, "space_id", "s", "", "space id")
        cmd.Flags().StringVarP(&interface_id, "interface_id", "i", "", "interface id")
        cmd.Flags().StringVarP(&kernel_version, "kernel_version", "k", "", "interface id")

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
