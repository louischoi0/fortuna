package swift

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
)

type PacketType uint8

const (
	PacketTypePing          PacketType = 0x00
	PacketTypePong          PacketType = 0x10
	PacketTypeErrorResponse PacketType = 0x0b

	PacketTypeGenVectorRequest     PacketType = 0x01
	PacketTypeStateSeedAPIRequest  PacketType = 0x02
	PacketTypeEmitEventRequest     PacketType = 0x03
	PacketTypeGenVectorResponse    PacketType = 0x04
	PacketTypeStateSeedAPIResponse PacketType = 0x05
	PacketTypeEmitEventResponse    PacketType = 0x06

	PacketTypeOracleGETSpaceRequest  PacketType = 0x07
	PacketTypeOracleGETSpaceResponse PacketType = 0x08

	PacketTypeReplicaConnectRequest  PacketType = 0x09
	PacketTypeReplicaConnectResponse PacketType = 0x0a

	PacketTypeGetOracleStatusRequest  PacketType = 0x20
	PacketTypeGetOracleStatusResponse PacketType = 0x21

	PacketTypePushPage PacketType = 0x22

	PacketTypeReplicaPageRequest  PacketType = 0x23
	PacketTypeReplicaPageResponse PacketType = 0x24

	PacketTypeGETUniverseInfoRequest  PacketType = 0x25
	PacketTypeGETUniverseInfoResponse PacketType = 0x26

	PacketTypeReplicaGetSpacePageNumRequest  PacketType = 0x27
	PacketTypeReplicaGetSpacePageNumResponse PacketType = 0x28



	// Replica Server API
	PacketTypeGetEventCountsRequest 	PacketType = 0x30
	PacketTypeGetEventCountsResponse 	PacketType = 0x3a

	PacketTypeGetEventDetailRequest  PacketType = 0x3b
	PacketTypeGetEventDetailResponse PacketType = 0x3c

	PacketTypeListEventsRequest  PacketType = 0x3d
	PacketTypeListEventsResponse PacketType = 0x3e
)

type Packet struct {
	Type    PacketType `json:"type"`
	Payload []byte     `json:"payload"`
	Header  string     `json:"header,omitempty"`
}

func ParseRemoteAddr(addr string) (string, string, string, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return "", "", "", fmt.Errorf("invalid address format: %w", err)
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = host[1 : len(host)-1]
	}

	ip := net.ParseIP(host)
	if ip == nil {
		return "", "", "", fmt.Errorf("invalid IP address: %s", host)
	}

	if ip.IsLoopback() {
		host = "localhost"
	}

	return fmt.Sprintf("%s:%s", host, port), host, port, nil
}

func FormatJSONResponse(payload []byte) string {
	var prettyJSON bytes.Buffer

	if err := json.Indent(&prettyJSON, payload, "", "    "); err != nil {
		log.Fatalf("JSON formatting failed: %v", err)
		return ""
	}
	return prettyJSON.String()
}

func AssertPacketType(packet *Packet, expected PacketType) error {
	if packet.Type != expected {
		log.Fatalf("expected packet type %d, but got %d", expected, packet.Type)
	}
	return nil
}

func NewGetUniverseInfoPacket() *Packet {
	return &Packet{
		Type:    PacketTypeGETUniverseInfoRequest,
		Payload: []byte(""),
	}
}

func NewReplicaPageRequest(spaceID string, pageNum int64) *Packet {
	return &Packet{
		Type:    PacketTypeReplicaPageRequest,
		Payload: []byte(fmt.Sprintf(`{"space_id": "%s", "page_num": %d}`, spaceID, pageNum)),
	}
}

func NewReplicaGetSpacePageNumRequest(spaceID string) *Packet {
	return &Packet{
		Type:    PacketTypeReplicaGetSpacePageNumRequest,
		Payload: []byte(fmt.Sprintf(`{"space_id": "%s"}`, spaceID)),
	}
}
