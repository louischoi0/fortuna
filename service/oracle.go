package service

import (
	"fortuna/core/model"
	"fortuna/swift"
	"sync"
)

const EVENT_BUFFER_SIZE = 1024
const TRANSACTION_BUFFER_SIZE = 1024

type Oracle struct {
	mu sync.Mutex

	// stateDB 		  *grocksdb.DB
	//ChainStorage		  *

	Chain			   *model.Chain
	CurrentBlock               *model.Block

	eventBuffer 		chan *model.Event
	transactionBuffer 	chan *model.Transaction

	swift       		*swift.TCPServer
}

type OracleConfig struct {
	MaxEventRequestsPerMinute int64
	EventBufferSize           int64
	TransactionBufferSize     int64
}


func NewOracle(config OracleConfig) *Oracle {
	/**
	stateDB, err := rock.GetDBInstance("status.rocks")
	if err != nil {
		log.Fatalf(err.Error())
	}
	**/
	return &Oracle{
		eventBuffer: make(chan *model.Event, EVENT_BUFFER_SIZE),
		transactionBuffer: make(chan *model.Transaction, TRANSACTION_BUFFER_SIZE),
	}
}

func (c *Oracle) Bootstrap() error {
	return nil
}

