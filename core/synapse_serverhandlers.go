package core

import (
	"encoding/json"
	"fortuna/core/vm"
	"fortuna/structure"
	"fortuna/core/model"
	"fortuna/core/util"
	"fortuna/swift"
	"context"
	"log"
)


func (syn *Synapse) StartUp() error {
        syn.swift.RegisterHandler(swift.PacketTypeStateSeedAPIRequest, func(ctx context.Context, packet *swift.Packet) error {
		omap, err := structure.ParseOrderedMap(string(packet.Payload))
		if err != nil {
			return syn.swift.SendErrorResponse(ctx, err.Error())
		}

		kernelVersion, ok := omap.String("kernel_version")
		if !ok {
			return syn.swift.SendErrorResponse(ctx, "payload must have key: kernel_version")
		}

		payload, ok := omap.String("payload")
		if !ok {
			return syn.swift.SendErrorResponse(ctx, "payload must have key: payload")
		}

		machine, err := syn.LoadMachine(vm.KernelVersion(kernelVersion))
		seed := machine.StateKernel.GenStateSeedPayload(payload)
		buffer, err := json.Marshal(seed)

		util.EncodeInt64Array([]int64{1,2,3})

		if err != nil {
			return syn.swift.SendErrorResponse(ctx, err.Error())
		}

                response := &swift.Packet{
                        Type:    swift.PacketTypeStateSeedAPIResponse,
                        Payload: json.RawMessage(buffer),
                }

                return syn.swift.Send(ctx, response)

	})

        syn.swift.RegisterHandler(swift.PacketTypeEmitEventRequest, func(ctx context.Context, packet *swift.Packet) error {
		omap, err := structure.ParseOrderedMap(string(packet.Payload))

		if err != nil {
			return syn.swift.SendErrorResponse(ctx, err.Error())
		}

		event, err := model.NewEventFromOrderedMap(omap)
		log.Printf("emit event requested. spec.KerenelVersion='%v' spec.InterfaceID='%v'", event.Spec.KernelVersion, event.Spec.InterfaceID)

		if err != nil {
			return syn.swift.SendErrorResponse(ctx, err.Error())
		}

		result, err := syn.Confirm(event)

		if err != nil {
			return syn.swift.SendErrorResponse(ctx, err.Error())
		}

		responseBuffer := result.Buffer()

                response := &swift.Packet{
                        Type:    swift.PacketTypeEmitEventResponse,
                        Payload: json.RawMessage(responseBuffer),
                }

                return syn.swift.Send(ctx, response)
        })

	return nil
}
