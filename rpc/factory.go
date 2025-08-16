package rpc

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"fortuna/core/model"
	"fortuna/swift"
	"io"
	"net"
	"time"
)

func NewRawRequest(peer string, packetType swift.PacketType, payload string) *RawRequest {
	return &RawRequest{
		Type:    packetType,
		Payload: payload,
		Peer:    peer,
	}
}

type RawRequest struct {
	Type    swift.PacketType
	Payload string
	Peer    string
	Auth    string
}

func (raw *RawRequest) A(auth string) *RawRequest {
	raw.Auth = auth
	return raw
}

func (req *RawRequest) Call() (swift.Packet, error) {
	packet := swift.Packet{
		Type:    req.Type,
		Payload: json.RawMessage(req.Payload),
	}

	return CallRPC(req.Peer, packet)
}

func CreateRequest(packetType swift.PacketType, peer, payload string) *RawRequest {
	req := RawRequest{
		Type:    packetType,
		Payload: payload,
		Peer:    peer,
	}

	return &req
}

func CreateEventRequest(peer string, event *model.Event, authorization string) *RawRequest {
	req := RawRequest{
		Type:    swift.PacketTypeEmitEventRequest,
		Payload: event.String(),
		Peer:    peer,
	}

	return &req
}

func CallRPC(targetNode string, packet swift.Packet) (swift.Packet, error) {
	nullpacket := swift.Packet{}
	conn, err := net.DialTimeout("tcp", targetNode, 3*time.Second)

	if err != nil {
		return nullpacket, fmt.Errorf("Failed to connect to server: %v", err)
	}

	defer conn.Close()

	packetData, err := json.Marshal(packet)
	if err != nil {
		return nullpacket, fmt.Errorf("Failed to serialize packet: %v", err)
	}

	packetLen := uint32(len(packetData))
	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header, packetLen)

	if _, err := conn.Write(header); err != nil {
		return nullpacket, fmt.Errorf("Failed to send header: %v", err)
	}
	if _, err := conn.Write(packetData); err != nil {
		return nullpacket, fmt.Errorf("Failed to send packet: %v", err)
	}

	if err := conn.SetReadDeadline(time.Now().Add(time.Second * 2)); err != nil {
		return nullpacket, fmt.Errorf("Failed to set read deadline: %v", err)
	}

	respHeader := make([]byte, 4)
	if _, err := io.ReadFull(conn, respHeader); err != nil {
		return nullpacket, fmt.Errorf("Failed to read response header: %v", err)
	}

	respPacketLen := binary.LittleEndian.Uint32(respHeader)

	respData := make([]byte, respPacketLen)
	if _, err := io.ReadFull(conn, respData); err != nil {
		return nullpacket, fmt.Errorf("Failed to read response packet: %v", err)
	}

	var response swift.Packet
	if err := json.Unmarshal(respData, &response); err != nil {
		return nullpacket, fmt.Errorf("Invalid response format: %v", err)
	}

	return response, nil
}

func CallRawRequest(request *RawRequest) (swift.Packet, error) {
	packet := swift.Packet{
		Type:    request.Type,
		Payload: json.RawMessage(request.Payload),
	}

	return CallRPC(request.Peer, packet)
}
