package service

import (
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

	swift    *swift.TCPServer
	synapses map[string]*component.Synapse
}

type OracleConfig struct {
	MaxEventRequestsPerMinute int64
	EventBufferSize           int64
	TransactionBufferSize     int64
}

func (o *Oracle) GetSynapse(spaceID string) *component.Synapse {
	syn, ok := o.synapses[spaceID]
	if !ok {
		return nil
	}

	return syn
}

func (o *Oracle) AllocateSpace(spaceID string) (*model.Space, error) {
	if _, exists := o.Universe[spaceID]; exists {
		return nil, fmt.Errorf("spaceID %s already exists", spaceID)
	}

	space := model.NewSpace(spaceID)
	o.Universe[spaceID] = space

	syn := component.NewSynapse(space)
	if err := syn.Bootstrap(); err != nil {
		return nil, fmt.Errorf("failed to bootstrap synapse: %v", err.Error())
	}

	o.synapses[spaceID] = syn

	return space, nil
}

func (o *Oracle) ListSpaces() ([]string, error) {
	spaces := make([]string, 0)
	for spaceID := range o.Universe {
		spaces = append(spaces, spaceID)
	}

	return spaces, nil
}

func GetOracleService(spaceID string, config OracleConfig) *Oracle {
	oracleOnce.Do(func() {
		universeDB, err := rock.GetDBInstance(fmt.Sprintf("meta.%s", spaceID))
		if err != nil {
			log.Fatalf("failed to get universe db: %v", err.Error())
		}

		if err != nil {
			log.Fatalf("failed to load space data: %v", err.Error())
		}

		swift := swift.NewServer()

		oracle = &Oracle{
			spaceID:           spaceID,
			eventBuffer:       make(chan *model.EventResult, EVENT_BUFFER_SIZE),
			transactionBuffer: make(chan *model.Transaction, TRANSACTION_BUFFER_SIZE),
			universe:          universeDB,
			swift:             swift,
			synapses:          make(map[string]*component.Synapse),
			Universe:          make(map[string]*model.Space),
		}

		_, err = oracle.AllocateSpace(spaceID)
		if err != nil {
			log.Fatalf("failed to allocate space: %v", err.Error())
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

func (o *Oracle) Daemon() error {
	log.Printf("oracle daemon started")

	for {
		select {
		case event := <-o.eventBuffer:
			log.Println("process event result buffer: ", event.GetSpaceID())

			space, err := o.GetSpace(event.GetSpaceID())
			if err != nil {
				log.Fatalf("failed to get space: %v", err.Error())
			}

			if space.CurrentPage == nil {
				log.Fatalf("space %s has no current page", event.GetSpaceID())
			}

			space.CurrentPage.AppendEventExecution(event)

			if o.ShouldCommitPage(space.CurrentPage) {
				err = o.CommitPage(space.CurrentPage)
				if err != nil {
					log.Fatalf("failed to commit page: %v", err.Error())
				}
			}
		}
	}
}

func (o *Oracle) ShouldCommitPage(page *model.Page) bool {
	log.Println("should commit page: ", page.GetCount())
	return page.GetCount() > 3
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
