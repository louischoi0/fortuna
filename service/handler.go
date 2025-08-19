package service

import (
	"context"
	"encoding/json"
	"fmt"
	"fortuna/core/model"
	"fortuna/core/vm"
	"fortuna/structure"
	"fortuna/swift"
	"fortuna/util"
	"log"
)

func (o *Oracle) RegisterHandlers() error {
	log.Printf("registering oracle handlers")

	o.swift.RegisterHandler(swift.PacketTypeReplicaConnectRequest, func(ctx context.Context, packet *swift.Packet) error {


		return nil
	})

	o.swift.RegisterHandler(swift.PacketTypeGenVectorRequest, func(ctx context.Context, packet *swift.Packet) error {
		omap, err := structure.ParseOrderedMap(string(packet.Payload))
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		kernelVersion, ok := omap.String("kernel_version")
		if !ok {
			return o.swift.SendErrorResponse(ctx, "payload must have key: kernel_version")
		}

		seed, ok := omap.String("seed")
		if !ok {
			return o.swift.SendErrorResponse(ctx, "payload must have key: seed")
		}

		size, ok := omap.Int64("size")
		if !ok {
			return o.swift.SendErrorResponse(ctx, "payload must have key: size")
		}

		spaceID, ok := omap.String("space_id")
		if !ok {
			return o.swift.SendErrorResponse(ctx, "payload must have key: space_id")
		}

		syn := o.GetSynapse(spaceID)
		machine, err := syn.LoadMachine(vm.KernelVersion(kernelVersion))
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		vector := machine.StateKernel.GenVector(seed, size)

		response := &swift.Packet{
			Type:    swift.PacketTypeGenVectorResponse,
			Payload: vector.Encode(),
		}

		return o.swift.Send(ctx, response)
	})

	o.swift.RegisterHandler(swift.PacketTypeStateSeedAPIRequest, func(ctx context.Context, packet *swift.Packet) error {
		return nil
	})

	o.swift.RegisterHandler(swift.PacketTypeOracleGETSpaceRequest, func(ctx context.Context, packet *swift.Packet) error {
		spaces, err := o.ListSpaces()
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		buffer, err := json.Marshal(spaces)
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		response := &swift.Packet{
			Type:    swift.PacketTypeOracleGETSpaceResponse,
			Payload: json.RawMessage(buffer),
		}

		return o.swift.Send(ctx, response)
	})

	o.swift.RegisterHandler(swift.PacketTypeStateSeedAPIRequest, func(ctx context.Context, packet *swift.Packet) error {
		omap, err := structure.ParseOrderedMap(string(packet.Payload))
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		kernelVersion, ok := omap.String("kernel_version")
		if !ok {
			return o.swift.SendErrorResponse(ctx, "payload must have key: kernel_version")
		}

		payload, ok := omap.String("payload")
		if !ok {
			return o.swift.SendErrorResponse(ctx, "payload must have key: payload")
		}

		spaceID, ok := omap.String("space_id")
		if !ok {
			return o.swift.SendErrorResponse(ctx, "payload must have key: space_id")
		}

		syn := o.GetSynapse(spaceID)
		machine, err := syn.LoadMachine(vm.KernelVersion(kernelVersion))
		seed := machine.StateKernel.GenStateSeedPayload(payload)
		buffer, err := json.Marshal(seed)

		util.EncodeInt64Array([]int64{1, 2, 3})

		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		response := &swift.Packet{
			Type:    swift.PacketTypeStateSeedAPIResponse,
			Payload: json.RawMessage(buffer),
		}

		return o.swift.Send(ctx, response)

	})

	o.swift.RegisterHandler(swift.PacketTypeEmitEventRequest, func(ctx context.Context, packet *swift.Packet) error {
		omap, err := structure.ParseOrderedMap(string(packet.Payload))

		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		event, err := model.NewEventFromOrderedMap(omap)
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		log.Printf("emit event requested. spec.KerenelVersion='%v' spec.InterfaceID='%v'", event.Spec.KernelVersion, event.Spec.InterfaceID)

		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		if space, err := o.GetSpace(event.SpaceID); err != nil || space == nil {
			return o.swift.SendErrorResponse(ctx, fmt.Sprintf("space '%s' not found", event.SpaceID))
		}

		syn := o.GetSynapse(event.SpaceID)
		result, err := syn.Confirm(event)
		o.eventBuffer <- result

		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		responseBuffer := result.Buffer()

		response := &swift.Packet{
			Type:    swift.PacketTypeEmitEventResponse,
			Payload: json.RawMessage(responseBuffer),
		}

		return o.swift.Send(ctx, response)
	})

	return nil
}

func (o *Oracle) Bootstrap() error {
	if err := o.RegisterHandlers(); err != nil {
		return err
	}

	return nil
}
