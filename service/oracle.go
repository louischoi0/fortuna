package service

import (
	"fortuna/core/model"
	"fortuna/core/storage"
	"fortuna/swift"
	"fortuna/rock"
	"sync"
)

const EVENT_BUFFER_SIZE = 1024
const TRANSACTION_BUFFER_SIZE = 1024

type Oracle struct {
	mu      		sync.Mutex

	storage 		*storage.FileStorage
	meta			*grocksdb.DB
	height			int64

	Chain        		*model.Chain
	CurrentBlock 		*model.Block

	eventBuffer       	chan *model.Event
	transactionBuffer 	chan *model.Transaction

	swift 			*swift.TCPServer
}

type OracleConfig struct {
	MaxEventRequestsPerMinute int64
	EventBufferSize           int64
	TransactionBufferSize     int64
}

func NewOracle(config OracleConfig) *Oracle {
	metaDB, err := rock.GetDBInstance("meta.rocks")
	if err != nil {
		log.Fatalf(err.Error())
	}

	return &Oracle{
		eventBuffer:       make(chan *model.Event, EVENT_BUFFER_SIZE),
		transactionBuffer: make(chan *model.Transaction, TRANSACTION_BUFFER_SIZE),
	}
}

func (o *Oracle) CommitCurentBlock() error {
	if c.CurrentBlock == nil {
		return fmt.Errorf("current block is not set (nil)")
	}

	current_block_buffer, err := c.CurrentBlock.Encode()
	if err != nil {
		return err
	}

	return nil
}

func (o *Oracle) SetHeight(height int64) error {
	err = rock.SetValue(sf.db, key, data)
	if err != nil {
		return err
	}
	o.height = height
}

func (c *Oracle) Bootstrap() error {
	return nil
}

func (c *Oracle) WriteBlock(block *model.Block) error {
	return nil
}
