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
	PacketTypePing				PacketType = 0x00
	PacketTypePong				PacketType = 0x10
	PacketTypeErrorResponse 		PacketType = 0x0b

	PacketTypeGenVectorRequest     		PacketType = 0x01
	PacketTypeStateSeedAPIRequest  		PacketType = 0x02
	PacketTypeEmitEventRequest     		PacketType = 0x03
	PacketTypeGenVectorResponse    		PacketType = 0x04
	PacketTypeStateSeedAPIResponse 		PacketType = 0x05
	PacketTypeEmitEventResponse    		PacketType = 0x06

	PacketTypeOracleGETSpaceRequest  	PacketType = 0x07
	PacketTypeOracleGETSpaceResponse 	PacketType = 0x08

	PacketTypeReplicaConnectRequest 	PacketType = 0x09
	PacketTypeReplicaConnectResponse 	PacketType = 0x0a

	PacketTypeGetOracleStatusRequest	PacketType = 0x20
	PacketTypeGetOracleStatusResponse	PacketType = 0x21

	PacketTypeSyncPage			PacketType = 0x22
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
