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
	ID            string
	mu            sync.Mutex
	metaDB        *grocksdb.DB
	Universe      map[string]*model.Space

	conn      net.Conn
	connected bool

	swift *swift.TCPServer
	dbs   map[string]*grocksdb.DB

	indexer *SpaceIndexer
}

func NewReplica() *Replica {
	log.Println("creating replica instance")
	swift := swift.NewServer()
	
	metaDB, err := rock.GetDBInstance(fmt.Sprintf("metastore.%s", "replica"))
        if err != nil {
        	log.Fatalf("failed to get space meta db: %v", err.Error())
        }

	return &Replica{
		swift:         	swift,
		metaDB:		metaDB,
		Universe:      	make(map[string]*model.Space),
		connected:     	false,
		dbs:           	make(map[string]*grocksdb.DB),
	}
}

func (rp *Replica) LoadUniverse() error {
	log.Println("load universe")
	spaces, err := model.ListSpaces(rp.metaDB)
	if err != nil {
		return err
	}

	for _, spaceID := range spaces {
		rp.Universe[spaceID] = model.NewSpace(spaceID, rp.metaDB)
	}
	return nil
}

func (rp *Replica) ActivateIndexer() *SpaceIndexer {
	log.Printf("activating replica event indexer %s", rp.ID)
	rp.indexer = NewSpaceIndexer(rp.Universe, rp.dbs)

	for _, space := range rp.Universe {
		rp.indexer.ScanPages(space.Pages)
	}

	return rp.indexer
}

func (rp *Replica) AllocateSpace(spaceID string) (*model.Space, error) {
	if _, exists := rp.Universe[spaceID]; exists {
		return nil, fmt.Errorf("spaceID %s already exists", spaceID)
	}

	space := model.NewSpace(spaceID, rp.metaDB)
	rp.Universe[spaceID] = space

	db, err := rock.GetDBInstance(fmt.Sprintf("meta.%s", spaceID))
	if err != nil {
		log.Fatalf("failed to get rocks db: %v", err.Error())
	}

	rp.dbs[spaceID] = db

	return space, nil
}

func (rp *Replica) GetSpaceLCNum(spaceID string) uint64 {
	space, ok := rp.Universe[spaceID]
	if !ok {
		return 0
	}
	return space.LastCommittedPageNum
}

func (rp *Replica) GetSpace(spaceID string) *model.Space {
	var err error
	space, ok := rp.Universe[spaceID]

	if !ok {
		space, err = rp.AllocateSpace(spaceID)
		if err != nil {
			log.Fatalf("replica has failed to allocate new space: ", err.Error())
		}
	}

	return space
}

func (rp *Replica) HandleBroadcastPage(page *model.Page) error {
	page.Update()

	space := rp.GetSpace(page.SpaceID)
	log.Printf("replica received page hash: %s, num: %d", page.Hash(), page.N)

	if page.GetPageNum() != space.LastCommittedPageNum + 1{
		log.Fatalf("page num does not matched expected %v, but %v", space.LastCommittedPageNum + 1, page.GetPageNum())
	}

	space.CurrentPage = page
	_, err := space.CommitCurrentPage()

	if rp.indexer != nil {
		rp.indexer.ScanPage(page)
	}
	
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

	err := rp.SyncAllSpaces(false)
	if err != nil {
		log.Fatalf("failed to sync all spaces before running replica service: %s", err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	ping := time.NewTicker(1 * time.Second)
	defer ping.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ping.C:
			err = rp.SyncAllSpaces(false)
			if err != nil {
				log.Fatalf("failed to sync: %s", err.Error())
			}
		}
	}
}

func (rp *Replica) LogUniverseInfo(info *UniverseInfo) {
	for spaceID, spaceInfo := range info.Spaces {
		log.Printf("master node space %s - hash: %s, page num: %d", spaceID, spaceInfo.Hash, spaceInfo.LastCommittedPageNum)
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

func (rp *Replica) CheckSpaceUptoDate(spaceID string) (bool, int64, error) {
	packet := swift.NewReplicaGetSpacePageNumRequest(spaceID)

	if err := rpc.SendPacket(rp.conn, packet); err != nil {
		return false, 0, err
	}

	response, err := rpc.ReadPacket(rp.conn)
	if err != nil {
		return false, 0, err
	}

	if err := swift.AssertPacketType(response, swift.PacketTypeReplicaGetSpacePageNumResponse); err != nil {
		log.Fatalf(err.Error())
	}

	var originPageNum int64
	err = json.Unmarshal(response.Payload, &originPageNum)
	if err != nil {
		return false, 0, err
	}

	lcn := rp.GetSpaceLCNum(spaceID)
	if int64(lcn) > originPageNum {
		log.Fatalf("replica page num higher than master")
	}

	return originPageNum == int64(lcn), originPageNum, nil
}

func (rp *Replica) SyncSpace(spaceID string, pageNum int64) error {

	for i := rp.GetSpaceLCNum(spaceID); int64(i) < pageNum; i++ {
		log.Printf("request space page for %s:%v", spaceID, i+1)
		packet := swift.NewReplicaPageRequest(spaceID, int64(i)+1)

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

		log.Println("page buffer size: ", len(response.Payload))
		page, err := model.DecodePage(response.Payload)
		if err != nil {
			log.Fatalf("invalid page received, failed to decod page: %s", err.Error())
			return err
		}

		log.Printf("page for space %s", page.SpaceID)
		err = rp.HandleBroadcastPage(page)
		if err != nil {
			return err
		}
	}

	return nil
}

func (rp *Replica) SyncAllSpaces(debugLogging bool) error {

	info, err := rp.GetOracleUniverseInfo()
	if err != nil {
		return err
	}

	if debugLogging{
		rp.LogUniverseInfo(info)
	}

	for spaceID, _ := range info.Spaces {
		space := rp.GetSpace(spaceID)
		if space == nil {
			log.Fatalf("space %s expected to exists", spaceID)
		}

		if debugLogging {
			log.Printf("start to sync space %s", spaceID)
		}

		upToDate, pageNum, err := rp.CheckSpaceUptoDate(spaceID)
		
		if debugLogging {
			log.Printf("master space pagenum: %v, uptodate: %v", pageNum, upToDate)
		}

		if err != nil {
			log.Fatalf(err.Error())
		} else if !upToDate {
			log.Printf("space %s is not up to date, start syncing from page %d to page %d", spaceID, space.LastCommittedPageNum, pageNum)
			err := rp.SyncSpace(spaceID, pageNum)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (rp *Replica) BootStrap() {
	log.Printf("registering replica handlers")

	rp.swift.RegisterHandler(swift.PacketTypeGetEventCountsRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
		if rp.indexer == nil {
			return rp.swift.SendErrorResponse(ctx, "replica indexer is not activated")
		}

		em := rp.indexer.GetEventCounts()
		buf, _ := json.Marshal(em)

		response := &swift.Packet{
			Type:    swift.PacketTypeGetEventCountsResponse,
			Payload: buf,
		}

		return rp.swift.Send(ctx, response)
	})

	rp.swift.RegisterHandler(swift.PacketTypeListEventsRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
		var req struct {
			SpaceID	string 	`json:"space_id"`
			Count	int	`json:"count"`	
		}

		err := json.Unmarshal(packet.Payload, &req)
		if err != nil {
			return rp.swift.SendErrorResponse(ctx, err.Error())
		}

		events := rp.indexer.ListEvents(req.SpaceID, 10)
		buf, _ := json.Marshal(events)

		response := &swift.Packet{
			Type:    swift.PacketTypeListEventsResponse,
			Payload: buf,
		}

		return rp.swift.Send(ctx, response)
	})

	rp.swift.RegisterHandler(swift.PacketTypeGetEventDetailRequest, func(ctx context.Context, conn net.Conn, packet *swift.Packet) error {
		// eventID := string(packet.Payload)
		// info := rp.indexer.GetExecutionInfo(eventID)

		return nil
	})

}
