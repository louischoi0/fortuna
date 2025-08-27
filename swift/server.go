package swift

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type PacketHandler func(ctx context.Context, conn net.Conn, packet *Packet) error

type TCPServer struct {
	handlers 	map[PacketType]PacketHandler
	peers    	map[string]net.Conn
	listener 	net.Listener

	host    	string
	port    	int
	address 	string

	onOpenCallbacks	[]func(net.Conn)

	mu sync.RWMutex
	c 		int64
}

func NewServer() *TCPServer {

	return &TCPServer{
		host:     "0.0.0.0",
		peers:    make(map[string]net.Conn),
		mu:       sync.RWMutex{},
		handlers: make(map[PacketType]PacketHandler),
	}
}

func (s *TCPServer) Start(port int) error {
	var err error
	s.address = fmt.Sprintf("%v:%v", s.host, port)

	s.listener, err = net.Listen("tcp", s.address)
	log.Printf("start swift server %v", s.address)

	if err != nil {
		return err
	}

	go s.acceptConnections()

	return nil
}

func (s *TCPServer) AddOnOpenCallback(cb func(net.Conn)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onOpenCallbacks = append(s.onOpenCallbacks, cb)
}

func (s *TCPServer) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			continue
		}

		for _, cb := range s.onOpenCallbacks {
			go cb(conn)
		}

		tcp, ok := conn.(*net.TCPConn)
		if ok {
			tcp.SetKeepAlive(true)
			tcp.SetKeepAlivePeriod(10 * time.Second)    // 10초마다 OS keep-alive 패킷
		}

		log.Println("swift tcp server connection opened")
		go s.HandleConnection(conn)
	}
}

func (s *TCPServer) HandleConnection(conn net.Conn) error {
	ctx := context.WithValue(context.Background(), "connection", conn)

	for {

		header := make([]byte, 4)
		if _, err := io.ReadFull(conn, header); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("failed to read header: %v", err)
		}


		packetLen := binary.LittleEndian.Uint32(header)
		packetBytes := make([]byte, packetLen)

		if _, err := io.ReadFull(conn, packetBytes); err != nil {
			return fmt.Errorf("failed to read packet: %v", err)
		}

		var packet Packet
		if err := json.Unmarshal(packetBytes, &packet); err != nil {
			return fmt.Errorf("failed to parse packet: %v", err)
		}

		switch packet.Type {
		case PacketTypePing:
			s.c = s.c + 1
			buf, _ := json.Marshal(s.c)

			if err := s.Send(ctx, &Packet{
				Type:    PacketTypePong,
				Payload: buf,
			}); err != nil {
				log.Println("send pong failed")
				return fmt.Errorf("failed to send pong response: %v", err)
			}
			continue
		default:
			if handler, ok := s.handlers[packet.Type]; ok {
				err := handler(ctx, conn, &packet)

				if err != nil {
					return s.SendErrorResponse(ctx, err.Error())
				}
			} else {
				return fmt.Errorf("unknown packet type: %v", packet.Type)
			}
		}
	}
}

func (s *TCPServer) Send(ctx context.Context, packet *Packet) error {
	connInfo, ok := ctx.Value("connection").(net.Conn)
	if !ok || connInfo == nil {
		return fmt.Errorf("connection info not found in context")
	}

	packetBytes, err := json.Marshal(packet)
	if err != nil {
		return fmt.Errorf("failed to serialize packet: %v", err)
	}

	header := make([]byte, 4)
	binary.LittleEndian.PutUint32(header, uint32(len(packetBytes)))

	if _, err := connInfo.Write(header); err != nil {
		return fmt.Errorf("failed to send header: %v", err)
	}
	if _, err := connInfo.Write(packetBytes); err != nil {
		return fmt.Errorf("failed to send packet: %v", err)
	}

	return nil
}

func (s *TCPServer) SendErrorResponse(ctx context.Context, errMsg string) error {
	errorPayload := struct {
		Error string `json:"error"`
	}{
		Error: errMsg,
	}

	payload, err := json.Marshal(errorPayload)
	if err != nil {
		return fmt.Errorf("error message serialization failed: %v", err)
	}
	return s.Send(ctx, &Packet{
		Type:    PacketTypeErrorResponse,
		Payload: payload,
	})
}

func (s *TCPServer) RegisterHandler(packetType PacketType, handler PacketHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[packetType] = handler
}
