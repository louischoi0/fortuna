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
	PacketTypeUnknown       PacketType = 0
	PacketTypePing          PacketType = 1
	PacketTypePong          PacketType = 2
	PacketTypeErrorResponse PacketType = 9

	PacketTypeEmitEventRequest  PacketType = 10
	PacketTypeEmitEventResponse PacketType = 11

	PacketTypeStateSeedAPIRequest  PacketType = 12
	PacketTypeStateSeedAPIResponse PacketType = 13

	PacketTypeGenVectorRequest  PacketType = 14
	PacketTypeGenVectorResponse PacketType = 15

	PacketTypeVerifyVectorRequest  PacketType = 16
	PacketTypeVerifyVectorResponse PacketType = 17
)

type Packet struct {
	Type    PacketType      `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Header  string          `json:"header,ommitempty"`
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

func FormatResponse(payload *json.RawMessage) string {
	var prettyJSON bytes.Buffer

	if err := json.Indent(&prettyJSON, *payload, "", "    "); err != nil {
		log.Fatalf("JSON formatting failed: %v", err)
		return ""
	}
	return prettyJSON.String()
}
