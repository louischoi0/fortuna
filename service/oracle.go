package service

import (
	"context"
	"fortuna/core/component"
	"fortuna/core/model"
	"fortuna/rock"
	"fortuna/swift"
	"log"
	"sync"

	"github.com/linxGnu/grocksdb"
)

const EVENT_BUFFER_SIZE = 1024
const TRANSACTION_BUFFER_SIZE = 1024

var oracleOnce sync.Once
var oracle *Oracle

type Oracle struct {
	mu sync.Mutex

	metaDB *grocksdb.DB

	Chain        *model.Chain
	currentBlock *model.Block

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

func GetOracleService(config OracleConfig) *Oracle {
	oracleOnce.Do(func() {
		metaDB, err := rock.GetDBInstance("meta.rocks")
		if err != nil {
			log.Fatalf(err.Error())
		}

		oracle = &Oracle{
			eventBuffer:       make(chan *model.EventResult, EVENT_BUFFER_SIZE),
			transactionBuffer: make(chan *model.Transaction, TRANSACTION_BUFFER_SIZE),
			metaDB:            metaDB,
		}
	})

	return oracle
}

func (o *Oracle) VerifyBlock(block *model.Block) error {
	return nil
}

func (o *Oracle) VerifyEventResult(event *model.EventResult) error {
	return nil
}

func (o *Oracle) ProcessEventResultBuffer(event *model.EventResult) error {
	for {
		select {
		case event := <-o.eventBuffer:
			o.currentBlock.AppendEventExecution(event)
		}
	}
}

func (o *Oracle) CommitBlock(block *model.Block) error {
	if err := o.VerifyBlock(block); err != nil {
		return err
	}

	err := o.Chain.CommitBlock(block)
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

func (o *Oracle) Run(port int) error {
	if err := o.swift.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err.Error())
	}

	return nil
}
