package core

import (
	"fortuna/structure"
)

func (syn *Synapse) StartUp() error {

        syn.swift.RegisterHandler(swift.PacketTypeEmitEventRequest, func(ctx context.Context, packet *swift.Packet) error {
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

                return o.swift.Send(ctx, response)
        })

}
