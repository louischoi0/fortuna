package component

import (
	"fortuna/core/model"
	"fortuna/swift"
	"sync"
)

type Collector struct {
	mu sync.Mutex

	MaxEventRequestsPerMinute int64
	CurretBlock               *model.EventBlock

	eventBuffer chan *model.Event
	swift       *swift.TCPServer
	throthled   bool
}

type CollectorConfig struct {
	MaxEventRequestsPerMinute int64
	EventBufferSize           int64
	TransactionBufferSize     int64
}

func NewCollector(config CollectorConfig) *Collector {
	return nil
}

func (c *Collector) Bootstrap() error {
	return nil
}
