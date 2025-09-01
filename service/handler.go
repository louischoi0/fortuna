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
	"net"
)

type PReplicaPageRequest struct {
	SpaceID string `json:"space_id"`
	PageNum int64  `json:"page_num"`
}

func (o *Oracle) RegisterHandlers() error {
	log.Printf("registering oracle handlers")

	o.swift.RegisterHandler(swift.PacketTypeReplicaConnectRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
		o.AddReplica(conn)

		response := &swift.Packet{
			Type:    swift.PacketTypeReplicaConnectResponse,
			Payload: []byte(""),
		}

		return o.swift.Send(ctx, response)
	})

	o.swift.RegisterHandler(swift.PacketTypeGetOracleStatusRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
		status := o.GetOracleStatus()
		buffer, err := json.Marshal(status)
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		response := &swift.Packet{
			Type:    swift.PacketTypeGetOracleStatusResponse,
			Payload: buffer,
		}

		return o.swift.Send(ctx, response)
	})

	o.swift.RegisterHandler(swift.PacketTypeGenVectorRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
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

	o.swift.RegisterHandler(swift.PacketTypeStateSeedAPIRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
		return nil
	})

	o.swift.RegisterHandler(swift.PacketTypeOracleGETSpaceRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
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

	o.swift.RegisterHandler(swift.PacketTypeStateSeedAPIRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
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

	o.swift.RegisterHandler(swift.PacketTypeEmitEventRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
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

	o.swift.RegisterHandler(swift.PacketTypeGETUniverseInfoRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
		info := o.GetUniverseInfo()
		buffer, err := json.Marshal(info)
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		response := &swift.Packet{
			Type:    swift.PacketTypeGETUniverseInfoResponse,
			Payload: buffer,
		}
		return o.swift.Send(ctx, response)
	})

	o.swift.RegisterHandler(swift.PacketTypeReplicaGetSpacePageNumRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
		var req struct {
			SpaceID string	`json:"space_id"`
		}

		err := json.Unmarshal(packet.Payload, &req)
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		space, err := o.GetSpace(req.SpaceID)

		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		npage := space.LastCommittedPageNum

		buf, err := json.Marshal(npage)
		if err != nil {
			return o.swift.SendErrorResponse(ctx, fmt.Sprintf("failed to ser pagenum %s", npage))
		}

		response := &swift.Packet{ Type: swift.PacketTypeReplicaGetSpacePageNumResponse, Payload: buf }

		return o.swift.Send(ctx, response)
	})


	o.swift.RegisterHandler(swift.PacketTypeReplicaPageRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
		var request PReplicaPageRequest
		err := json.Unmarshal(packet.Payload, &request)
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		space, err := o.GetSpace(request.SpaceID)
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		page, err := space.GetPage(uint64(request.PageNum))
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		buffer, err := page.Encode()
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}
		log.Printf("oracle broad cast page %s, previoushash: %s, num: %d, buffer size: %d", page.Hash(), page.PrevPageHash, page.N,  len(buffer))

		response := &swift.Packet{
			Type:    swift.PacketTypeReplicaPageResponse,
			Payload: buffer,
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
