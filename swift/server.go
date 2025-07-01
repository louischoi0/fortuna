package swift

import (
	"sync"
	"net"
)


type PacketHandler func(ctx context.Context, packet *Packet) error

type TCPServer struct {
	handlers      map[PacketType]func(ctx context.Context, packet *Packet) error
	peers         map[string]net.Conn
	listener      net.Listener
	mu            sync.RWMutex
}

func (s *TCPServer) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}
		go s.HandleConnection(conn)
	}
}



