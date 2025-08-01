package component
	
import (
	"fortuna/core/model"
	"sync"
)

type Collector struct {
	mu 				sync.Mutext

	MaxEventRequestsPerMinute 	int64
	CurretBlock			*model.EventBlock

	eventBuffer			chan *Event
	swift 				*swift.TCPServer
	throthled			bool
}

type CollectorConfig struct {
	MaxEventRequestsPerMinute 	int64
	EventBufferSize			int64
	TransactionBufferSize 		int64
}

func NewCollector(config CollectorConfig) *Collector {
	return nil
}

func (c *Collector) Bootstrap() error {
	return nil
}


