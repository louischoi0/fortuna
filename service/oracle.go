package service

import (
	"context"
	"fmt"
	"fortuna/core/component"
	"fortuna/core/model"
	"fortuna/rock"
	"fortuna/structure"
	"fortuna/swift"
	"log"
	"sync"

	"os"
	"os/signal"
	"syscall"

	"github.com/linxGnu/grocksdb"
)

const EVENT_BUFFER_SIZE = 1024
const TRANSACTION_BUFFER_SIZE = 1024

var oracleOnce sync.Once
var oracle *Oracle

type Oracle struct {
	mu sync.Mutex

	spaceID  string
	universe *grocksdb.DB

	Universe map[string]*model.Space

	priorityQueue *structure.PriorityQueue

	eventBuffer       chan *model.EventResult
	transactionBuffer chan *model.Transaction

	swift *swift.TCPServer
	syn   *component.Synapse
}

type OracleConfig struct {
	MaxEventRequestsPerMinute int64
	EventBufferSize           int64
	TransactionBufferSize     int64
}

func (o *Oracle) AllocateSpace(spaceID string) error {
	if _, exists := o.Universe[spaceID]; exists {
		return fmt.Errorf("spaceID %s already exists", spaceID)
	}

	space := model.NewSpace(spaceID)
	o.Universe[spaceID] = space

	return nil
}

func GetOracleService(spaceID string, config OracleConfig) *Oracle {
	oracleOnce.Do(func() {
		universeDB, err := rock.GetDBInstance(fmt.Sprintf("meta.%s", spaceID))
		if err != nil {
			log.Fatalf("failed to get universe db: %v", err.Error())
		}

		space := model.NewSpace(spaceID)
		err = space.LoadSpaceData()

		if err != nil {
			log.Fatalf("failed to load space data: %v", err.Error())
		}

		swift := swift.NewServer()
		syn := component.NewSynapse(space)

		oracle = &Oracle{
			spaceID:           spaceID,
			eventBuffer:       make(chan *model.EventResult, EVENT_BUFFER_SIZE),
			transactionBuffer: make(chan *model.Transaction, TRANSACTION_BUFFER_SIZE),
			universe:          universeDB,
			swift:             swift,
			syn:               syn,
			Universe:          make(map[string]*model.Space),
		}
	})

	return oracle
}

func (o *Oracle) VerifyPage(page *model.Page) error {
	return nil
}

func (o *Oracle) VerifyEventResult(event *model.EventResult) error {
	return nil
}

func (o *Oracle) ProcessEventResultBuffer(event *model.EventResult) error {
	for {
		select {
		case event := <-o.eventBuffer:
			space, err := o.GetSpace(event.GetSpaceID())
			if err != nil {
				return err
			}

			if space.LastPage == nil {
				return fmt.Errorf("space %s has no last page", event.GetSpaceID())
			}

			space.LastPage.AppendEventExecution(event)
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

func (o *Oracle) CommitPage(page *model.Page) error {

	if err := o.VerifyPage(page); err != nil {
		return err
	}

	space := o.Universe[o.spaceID]

	err := space.CommitPage(page)
	if err != nil {
		return err
	}

	return nil
}

func (c *Oracle) Bootstrap() error {
	return nil
}

const (
	PacketTypeConfirmEventRequest  swift.PacketType = 101
	PacketTypeConfirmEventResponse swift.PacketType = 102
)

func (o *Oracle) RegisterHandlers() error {

	o.swift.RegisterHandler(PacketTypeConfirmEventRequest, func(ctx context.Context, packet *swift.Packet) error {
		er, err := model.DecodeEventResult(packet.Payload)
		if err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		if err := o.VerifyEventResult(er); err != nil {
			return o.swift.SendErrorResponse(ctx, err.Error())
		}

		o.eventBuffer <- er
		return nil
	})

	o.swift.RegisterHandler(swift.PacketTypePing, func(ctx context.Context, packet *swift.Packet) error {
		return o.swift.Send(ctx, &swift.Packet{
			Type:    swift.PacketTypePong,
			Payload: []byte("pong"),
		})
	})

	return nil
}

func (o *Oracle) Shutdown() error {
	for _, s := range o.Universe {
		s.Storage.Close()
	}

	return nil
}

func (o *Oracle) Run(port int) error {
	if err := o.syn.Run(); err != nil {
		log.Fatalf("Failed to start synapse: %v", err.Error())
	}

	if err := o.swift.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err.Error())
	}

	if err := o.RegisterHandlers(); err != nil {
		log.Fatalf("Failed to register handlers: %v", err.Error())
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	o.Shutdown()

	return nil
}
