package core

import (
	"encoding/json"
	"fortuna/structure"
	"fortuna/core/model"
	"fortuna/swift"
	"context"
	"log"
)

func (syn *Synapse) StartUp() error {

        syn.swift.RegisterHandler(swift.PacketTypeEmitEventRequest, func(ctx context.Context, packet *swift.Packet) error {
		log.Println("emit event requested")
		omap, err := structure.ParseOrderedMap(string(packet.Payload))

		if err != nil {
			return err
		}

		event, err := model.NewEventFromOrderedMap(omap)

		if err != nil {
			return err
		}
		
		result, err := syn.Confirm(event)
		responseBuffer := result.Buffer()

                response := &swift.Packet{
                        Type:    swift.PacketTypeEmitEventResponse,
                        Payload: json.RawMessage(responseBuffer),
                }

                return syn.swift.Send(ctx, response)
        })

	return nil
}
