package component

import (
	"fortuna/core/model"
	"fortuna/swift"
	"sync"
)

const COLLECTOR_EVENT_BUFFER_SIZE = 1024
const COLLECTOR_TRANSACTION_BUFFER_SIZE = 1024

type Collector struct {
	mu sync.Mutex

	// stateDB 		  *grocksdb.DB
	//ChainStorage		  *

	Chain        *model.Chain
	CurrentBlock *model.Block

	eventBuffer       chan *model.Event
	transactionBuffer chan *model.Transaction

	swift *swift.TCPServer
}

type CollectorConfig struct {
	MaxEventRequestsPerMinute int64
	EventBufferSize           int64
	TransactionBufferSize     int64
}

func NewCollector(config CollectorConfig) *Collector {
	/**
	stateDB, err := rock.GetDBInstance("status.rocks")
	if err != nil {
		log.Fatalf(err.Error())
	}
	**/
	return &Collector{
		eventBuffer:       make(chan *model.Event, COLLECTOR_EVENT_BUFFER_SIZE),
		transactionBuffer: make(chan *model.Transaction, COLLECTOR_TRANSACTION_BUFFER_SIZE),
	}
}

func (c *Collector) Bootstrap() error {
	return nil
}

func (c *Collector) CommitBlock() error {
	return nil
}

func (c *Collector) VerifyEventExecution(event *model.EventExecutionResult) error {
	return nil
}
