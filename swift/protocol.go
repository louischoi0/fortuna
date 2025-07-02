package swift

import (
	"fmt"
	"net"
	"strings"
	"encoding/json"
)

type PacketType uint8

const (
	PacketTypeUnknown		PacketType = 0
	PacketTypePing			PacketType = 1
	PacketTypePong			PacketType = 2
	PacketTypeErrorResponse		PacketType = 9
)

type Packet struct {
	Type    	PacketType      `json:"type"`
	Payload 	json.RawMessage `json:"payload"`
	Header		string		`json:"header,ommitempty"`
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
