package service

import (
	"fmt"
	"fortuna/core/component"
	"fortuna/core/model"
	"fortuna/rock"
	"fortuna/rpc"
	"fortuna/swift"
	"fortuna/util"
	"log"
	"sync"

	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/linxGnu/grocksdb"
)

const EVENT_BUFFER_SIZE = 1024
const TRANSACTION_BUFFER_SIZE = 1024

var oracleOnce sync.Once
var oracle *Oracle

type replicaInfo struct {
	ID   string `json:"id"`
	conn net.Conn
}

func NewReplicaInfo(rid string, conn net.Conn) replicaInfo {
	return replicaInfo{
		ID:   rid,
		conn: conn,
	}
}

type Oracle struct {
	mu sync.Mutex

	dbs    map[string]*grocksdb.DB
	metaDB *grocksdb.DB

	Universe      map[string]*model.Space

	eventBuffer       chan *model.EventResult
	transactionBuffer chan *model.Transaction
	pageSignal        chan *model.Page

	swift    *swift.TCPServer
	synapses map[string]*component.Synapse

	replicas map[string]replicaInfo
}

type UniverseInfo struct {
	Spaces map[string]model.SpaceInfo
}

type OracleConfig struct {
	MaxEventRequestsPerMinute int64
	EventBufferSize           int64
	TransactionBufferSize     int64
}

type OracleStatus struct {
	Status  string
	Version string
	Replica map[string]replicaInfo `json:"replica"`
}

func (o *Oracle) GetUniverseInfo() *UniverseInfo {
	res := make(map[string]model.SpaceInfo)

	for space_id, space := range o.Universe {
		res[space_id] = space.Info()
	}

	return &UniverseInfo{Spaces: res}
}

func (o *Oracle) GetOracleStatus() OracleStatus {
	var res OracleStatus

	res.Replica = o.replicas
	res.Status = "running"
	res.Version = "v1.0.0"

	return res
}

func (o *Oracle) GetSynapse(spaceID string) *component.Synapse {
	syn, ok := o.synapses[spaceID]
	if !ok {
		return nil
	}

	return syn
}

func (o *Oracle) AllocateSpace(spaceID string, metaDB *grocksdb.DB) (*model.Space, error) {
	if _, exists := o.Universe[spaceID]; exists {
		return nil, fmt.Errorf("spaceID %s already exists", spaceID)
	}

	space := model.NewSpace(spaceID, metaDB)
	o.Universe[spaceID] = space

	syn := component.NewSynapse(space)
	if err := syn.Bootstrap(); err != nil {
		return nil, fmt.Errorf("failed to bootstrap synapse: %v", err.Error())
	}

	db, err := rock.GetDBInstance(fmt.Sprintf("meta.%s", spaceID))
	if err != nil {
		log.Fatalf("failed to get rocks db: %v", err.Error())
	}

	o.synapses[spaceID] = syn
	o.dbs[spaceID] = db

	return space, nil
}

func (o *Oracle) ListSpaces() ([]string, error) {
	spaces := make([]string, 0)
	for spaceID := range o.Universe {
		spaces = append(spaces, spaceID)
	}

	return spaces, nil
}

func GetOracleService(oracleNodeID string, initialSpaceID string, config OracleConfig) *Oracle {
	oracleOnce.Do(func() {
		swift := swift.NewServer()
		metaDB, err := rock.GetDBInstance(fmt.Sprintf("metastore.%s", oracleNodeID))

		if err != nil {
			log.Fatalf("failed to get space meta db: %v", err.Error())
		}

		oracle = &Oracle{
			eventBuffer:       make(chan *model.EventResult, EVENT_BUFFER_SIZE),
			transactionBuffer: make(chan *model.Transaction, TRANSACTION_BUFFER_SIZE),
			dbs:               make(map[string]*grocksdb.DB),
			metaDB:            metaDB,
			swift:             swift,
			synapses:          make(map[string]*component.Synapse),
			Universe:          make(map[string]*model.Space),
			replicas:          make(map[string]replicaInfo),
		}

		_, err = oracle.AllocateSpace(initialSpaceID, metaDB)
		if err != nil {
			log.Fatalf("failed to allocate space: %v", err.Error())
		}
	})

	return oracle
}

func (o *Oracle) LoadUniverse() error {
	spaces, err := model.ListSpaces(o.metaDB)
	if err != nil {
		return err
	}

	for _, spaceID := range spaces {
		if _, exists := o.Universe[spaceID]; exists {
			continue
		}
		o.Universe[spaceID] = model.NewSpace(spaceID, o.metaDB)
	}
	return nil
}

func (o *Oracle) NewPushPagePacket(page *model.Page) *swift.Packet {
	buf, err := page.Encode()
	if err != nil {
		log.Fatalf(err.Error())
	}

	return &swift.Packet{
		Type:    swift.PacketTypePushPage,
		Payload: buf,
	}
}

func (o *Oracle) PushPageReplica(replica replicaInfo, page *model.Page) error {
	packet := o.NewPushPagePacket(page)
	return rpc.SendPacket(replica.conn, packet)
}

func (o *Oracle) FallbackPushPage(replica replicaInfo, page *model.Page) error {
	return nil
}

func (o *Oracle) PushPageReplicas(page *model.Page) error {
	for _, replica := range o.replicas {
		err := o.PushPageReplica(replica, page)
		if err != nil {
			o.FallbackPushPage(replica, page)
		}
	}

	return nil
}

func (o *Oracle) HandleEventResultBuffer(event *model.EventResult) error {
	log.Println("process event result buffer: ", event.GetSpaceID())
	spaceID := event.GetSpaceID()
	space, err := o.GetSpace(spaceID)
	if err != nil {
		log.Fatalf("failed to get space: %v", err.Error())
	}

	if space.CurrentPage == nil {
		log.Fatalf("space %s has no current page", event.GetSpaceID())
	}

	space.CurrentPage.AppendEventExecution(event)

	npage, err := space.MaybeCommitPage()
	if err != nil {
		log.Fatalf("failed to commit page: %v", err.Error())
	}

	if npage == nil {
		return nil
	}

	log.Println("set space page num map for space %s = %v", spaceID, uint64(space.LastCommittedPageNum))
	o.pageSignal <- npage

	return nil
}

func (o *Oracle) Daemon() error {
	log.Printf("oracle daemon started")

	for {
		select {
		case event := <-o.eventBuffer:
			go o.HandleEventResultBuffer(event)

		case npage := <-o.pageSignal:
			log.Println("todo broadcast page to replicas")
			o.PushPageReplicas(npage)
		}
	}
}

func (o *Oracle) GetSpace(spaceID string) (*model.Space, error) {
	u, ok := o.Universe[spaceID]
	if !ok {
		return nil, fmt.Errorf("spaceID %s not found", spaceID)
	}

	return u, nil
}

func (o *Oracle) Shutdown() error {
	for _, s := range o.Universe {
		s.Storage.Close()
	}

	return nil
}

func (o *Oracle) Run(port int) error {

	if err := o.swift.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err.Error())
	}

	if err := o.Bootstrap(); err != nil {
		log.Fatalf("Failed to bootstrap: %v", err.Error())
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go o.Daemon()

	<-sigChan

	o.Shutdown()

	return nil
}

func (o *Oracle) AddReplica(conn net.Conn) {
	s, _ := util.RandomBase64(6)
	rpi := NewReplicaInfo(s, conn)
	o.replicas[rpi.ID] = rpi
}
