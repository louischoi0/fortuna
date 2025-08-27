package service

import (
	"context"
	"encoding/json"
	"fmt"
	"fortuna/core/model"
	"fortuna/rock"
	"fortuna/rpc"
	"fortuna/swift"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/linxGnu/grocksdb"
)

type Replica struct {
	mu            sync.Mutex
	universeDB    *grocksdb.DB
	Universe      map[string]*model.Space
	SpacePageNums map[string]int64

	conn      net.Conn
	connected bool

	swift *swift.TCPServer
	dbs   map[string]*grocksdb.DB
}

func NewReplica() *Replica {
	swift := swift.NewServer()
	return &Replica{
		swift:         swift,
		Universe:      make(map[string]*model.Space),
		connected:     false,
		dbs:           make(map[string]*grocksdb.DB),
		SpacePageNums: make(map[string]int64),
	}
}

func (rp *Replica) AllocateSpace(spaceID string) (*model.Space, error) {
	if _, exists := rp.Universe[spaceID]; exists {
		return nil, fmt.Errorf("spaceID %s already exists", spaceID)
	}

	space := model.NewSpace(spaceID)
	rp.Universe[spaceID] = space

	db, err := rock.GetDBInstance(fmt.Sprintf("meta.%s", spaceID))
	if err != nil {
		log.Fatalf("failed to get rocks db: %v", err.Error())
	}

	rp.dbs[spaceID] = db
	rp.SpacePageNums[spaceID] = 0

	return space, nil
}

func (rp *Replica) GetSpace(spaceID string) *model.Space {
	space, ok := rp.Universe[spaceID]
	if !ok {
		return nil
	}
	return space
}

func (rp *Replica) HandleBroadcastPage(page *model.Page) error {
	space := rp.GetSpace(page.SpaceID)
	space.CurrentPage = page
	_, err := space.CommitCurrentPage()
	rp.SpacePageNums[page.SpaceID] = int64(page.GetPageNum())
	return err
}

func (rp *Replica) Connect(addr string) error {
	conn, err := rpc.Connect(addr)

	if err != nil {
		return err
	}

	rp.conn = conn
	packet := rpc.NewReplicaHandshakePacket()

	if err := rpc.SendPacket(rp.conn, packet); err != nil {
		return rp.Shutdown(err.Error())
	}

	response, err := rpc.ReadPacket(rp.conn)
	if err != nil {
		return rp.Shutdown(err.Error())
	}

	if response.Type != swift.PacketTypeReplicaConnectResponse {
		log.Fatalf("invalid packet received, expected %s but, %s", swift.PacketTypeReplicaConnectResponse, response.Type)
	}

	log.Println("successfully connected and a handshake done")
	rp.conn = conn
	rp.connected = true

	return nil
}

func (rp *Replica) Shutdown(msg string) error {
	log.Println("shutdown replica cause: %s", msg)
	return nil
}

func (rp *Replica) HandleRecvPacket(packet *swift.Packet) error {

	switch packet.Type {
	case swift.PacketTypePong:
		log.Println(string(packet.Payload))
	case swift.PacketTypeReplicaPageRequest:
		page, err := model.DecodePage(packet.Payload)
		if err != nil {
			log.Fatalf("invalid page buffer, failed to decode page: %v", err.Error())
		}

		err = rp.HandleBroadcastPage(page)
		if err != nil {
			log.Fatalf("failed to handle broadcast page: %v", err.Error())
		}
	}

	return nil
}

func (rp *Replica) PingMasterServer() error {
	pingPacket := rpc.NewPingPacket()

	pingPacket.Payload, _ = json.Marshal("pong")

	if err := rpc.SendPacket(rp.conn, pingPacket); err != nil {
		return rp.Shutdown(err.Error())
	}

	packet, err := rpc.ReadPacket(rp.conn)
	if err != nil {
		log.Println("failed to read packet")
		return rp.Shutdown(err.Error())
	}

	return rp.HandleRecvPacket(packet)
}

func (rp *Replica) Run() error {
	if !rp.connected {
		log.Fatalf("replica not connected")
	}

	info, err := rp.GetOracleUniverseInfo()
	if err != nil {
		return err
	}

	rp.LogUniverseInfo(info)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ping := time.NewTicker(1 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (rp *Replica) LogUniverseInfo(info *UniverseInfo) {
	for spaceID, spaceInfo := range info.Spaces {
		log.Printf("space %s - hash: %s, page num: %d", spaceID, spaceInfo.Hash, spaceInfo.PageNum)
	}
}

func (rp *Replica) GetOracleUniverseInfo() (*UniverseInfo, error) {
	packet := swift.NewGetUniverseInfoPacket()

	if err := rpc.SendPacket(rp.conn, packet); err != nil {
		return nil, err
	}

	response, err := rpc.ReadPacket(rp.conn)
	if err != nil {
		return nil, err
	}

	if err := swift.AssertPacketType(response, swift.PacketTypeGETUniverseInfoResponse); err != nil {
		return nil, err
	}

	var info UniverseInfo
	err = json.Unmarshal(response.Payload, &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

func (rp *Replica) CheckSpaceUptoDate(spaceID string, date time.Time) (bool, int64, error) {
	packet := swift.NewReplicaGetSpacePageNumRequest(spaceID)

	if err := rpc.SendPacket(rp.conn, packet); err != nil {
		return false, 0, err
	}

	response, err := rpc.ReadPacket(rp.conn)
	if err != nil {
		return false, 0, err
	}

	if err := swift.AssertPacketType(response, swift.PacketTypeReplicaGetSpacePageNumResponse); err != nil {
		return false, 0, err
	}

	var request int64
	err = json.Unmarshal(response.Payload, &request)
	if err != nil {
		return false, 0, err
	}

	return request == rp.SpacePageNums[spaceID], request, nil
}

func (rp *Replica) SyncSpace(spaceID string, pageNum int64) error {
	for i := rp.SpacePageNums[spaceID]; i < pageNum; i++ {
		packet := swift.NewReplicaPageRequest(spaceID, i)
		if err := rpc.SendPacket(rp.conn, packet); err != nil {
			return err
		}

		response, err := rpc.ReadPacket(rp.conn)
		if err != nil {
			return err
		}

		if err := swift.AssertPacketType(response, swift.PacketTypeReplicaPageResponse); err != nil {
			return err
		}

		page, err := model.DecodePage(response.Payload)
		if err != nil {
			return err
		}

		err = rp.HandleBroadcastPage(page)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rp *Replica) SyncAllSpaces() error {
	for spaceID, _ := range rp.Universe {
		upToDate, pageNum, err := rp.CheckSpaceUptoDate(spaceID, time.Now())
		if err != nil {
			return err
		} else if !upToDate {
			log.Printf("space %s is not up to date, start syncing to page %d", spaceID, pageNum)
			err := rp.SyncSpace(spaceID, pageNum)
			if err != nil {
				return err
			}
		} else {
			log.Printf("space %s is up to date", spaceID)
		}
	}

	return nil
}
