package service

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/core/vm"
	"fortuna/structure"

	"fortuna/rock"
	"fortuna/util"
	"log"
	"sync"

	"github.com/linxGnu/grocksdb"
)


type MachineStateSnapshot struct {
	MachineID   	string
	SpaceID     	string
	ActivatedAt 	uint64
	Duration    	int64
	State	    	*model.StateVector
}

type MachineStateHistory struct {
	MachineID	string
	SpaceID		string
	history		*structure.SortedList[uint64, *MachineStateSnapshot]
}

func NewMachineStateHistory(spaceID, machineID string) *MachineStateHistory {
	log.Printf("create new machine state hist instance for %s", machineID)

	return &MachineStateHistory{
		SpaceID: spaceID,
		MachineID: machineID,
		history: structure.NewSortedList[uint64, *MachineStateSnapshot](),
	}
}

func (ts *MachineStateHistory) AddMachineStateLogTransaction(tx *model.Transaction) {
	spaceID, machineID, state := vm.GetMachineStateFromTransaction(tx)
	log.Printf("add machine state log transaction %s", tx.Hash())

	sn := &MachineStateSnapshot{
		MachineID: machineID,
		SpaceID: spaceID,
		State: state,
		ActivatedAt: tx.Timestamp,
	}

	ts.history.Set(tx.Timestamp, sn)
}

type EventIndex struct {
	EventID string
	SpaceID string
	PageNum uint64
	Length  uint64
	Hash    string
}

type SpaceIndexer struct {
	mu        	sync.Mutex
	Universe  	map[string]*model.Space
	dbs       	map[string]*grocksdb.DB
	indexerDB 	*grocksdb.DB

	eventIndicies 			map[string]map[string]*EventIndex
	events        		map[string]map[string]*model.EventResult

	// EventID <-> SpaceID
	eventSpaceIDMappingCache	map[string]string
	eventCache			map[string]*model.EventResult

	eventCounts			map[string]int64
	//TODO snapshots

	// SpaceID -> machineID -> ts
	stateHistory map[string]map[string]*MachineStateHistory
}

func (si *SpaceIndexer) GetMachineStateHistory(spaceID string, machineID string) *MachineStateHistory {
	machines, ok := si.stateHistory[spaceID]
	if !ok {
		machines = make(map[string]*MachineStateHistory)
		si.stateHistory[spaceID] = machines
	}

	history, ok := machines[machineID]
	if !ok {
		history = NewMachineStateHistory(spaceID, machineID)
		machines[machineID] = history
	}

	return history

}

func (si *SpaceIndexer) GetExecutionInfo(eventID string) (*model.EventResult, error) {
	ex, ok := si.eventCache[eventID]
	if !ok {
		return nil, fmt.Errorf("event %s not found", eventID)
	}
	return ex, nil
}

func NewSpaceIndexer(universe map[string]*model.Space, dbs map[string]*grocksdb.DB) *SpaceIndexer {
	log.Printf("create space indexer instance")

	db, err := rock.GetDBInstance("indexer.meta")
	if err != nil {
		log.Fatalf("failed to get rocks db: %v", err.Error())
	}

	return &SpaceIndexer{
		Universe:      universe,
		dbs:           dbs,
		indexerDB:     db,
		eventIndicies: make(map[string]map[string]*EventIndex),
		events:        make(map[string]map[string]*model.EventResult),
		eventCounts:   make(map[string]int64),
		eventSpaceIDMappingCache: make(map[string]string),
		eventCache: make(map[string]*model.EventResult),
		stateHistory: make(map[string]map[string]*MachineStateHistory),
	}
}

type MachineStateLog struct {
	SpaceID		string
	MachineID	string
	Timestamp	uint64
	State		*model.StateVector
}

func FormatStateLogTransactionKey(spaceID, machineID string, timestamp uint64) string {
	return fmt.Sprintf("sv:%s:%s:%d", spaceID, machineID, timestamp)
}

func (si *SpaceIndexer) GetMachineStateLog(spaceID string, machineID string, timestamp uint64) (*MachineStateLog, error) {
	k := FormatStateLogTransactionKey(spaceID, machineID, timestamp)
	buf, err := si.ReadIndexerDB(k)
	if err != nil {
		return nil, err
	}

	sv, err := model.DecodeStateVector(buf)	
	if err != nil {
		return nil, err
	}
	res := &MachineStateLog{ State: sv, SpaceID: spaceID, MachineID: machineID, Timestamp: timestamp }
	return res, nil
}

func (si *SpaceIndexer) FormatKeyStateLogTransaction(tx *model.Transaction) string {
	spaceID, machineID, _ := vm.GetMachineStateFromTransaction(tx)
	return fmt.Sprintf("tx:%s:%s:%d", spaceID, machineID, tx.Timestamp)
}

func (si *SpaceIndexer) WriteIndexerDB(k string, v []byte) error {
	return rock.SetValue(si.indexerDB, k, v)
}

func (si *SpaceIndexer) ReadIndexerDB(k string) ([]byte, error) {
	return rock.GetValue(si.indexerDB, k)
}

func (si *SpaceIndexer) ReadEvent(eventID string) (*model.EventResult, error) {
	return nil, nil
}

func (si *SpaceIndexer) GetSpace(spaceID string) (*model.Space, error) {
	space, ok := si.Universe[spaceID]
	if !ok {
		return nil, fmt.Errorf("space %s not found", spaceID)
	}

	return space, nil
}

func (si *SpaceIndexer) ReadPage(spaceID string, pageNum int64) (*model.Page, error) {
	si.mu.Lock()
	defer si.mu.Unlock()

	space, err := si.GetSpace(spaceID)
	if err != nil {
		return nil, err
	}

	page, err := space.ReadPage(uint64(pageNum))
	if err != nil {
		return nil, err
	}

	return page, nil
}

func (si *SpaceIndexer) ScanPages(pages []*model.Page) error {

	for _, page := range pages {
		/**
		lcn, _ := si.GetLastIndexedPageNum(page.SpaceID)

		if page.GetPageNum() <= lcn {
			continue
		}
		**/

		err := si.ScanPage(page)
		if err != nil {
			return err
		}
	}
	return nil
}

func (si *SpaceIndexer) SetLastIndexedPageNum(spaceID string, pageNum uint64) error {
	return rock.SetValue(si.indexerDB, "last_indexed_page", util.EncodeUint64(pageNum))
}

func (si *SpaceIndexer) GetLastIndexedPageNum(spaceID string) (uint64, error) {
	buf, err := rock.GetValue(si.indexerDB, "last_indexed_page")
	if err != nil {
		return 0, err
	}
	
	i, err := util.DecodeUint64(buf)

	if err != nil {
		return 0, err
	}
	
	return i, nil
}


func (si *SpaceIndexer) AddExecution(pageNum uint64, execution *model.EventResult) {
	eventID := execution.Hash()

	spaceID := execution.Event.SpaceID

	if _, ok := si.eventIndicies[spaceID]; !ok{
		si.eventIndicies[spaceID] = make(map[string]*EventIndex)
	}

	if _, ok := si.events[spaceID]; !ok {
		si.events[spaceID] = make(map[string]*model.EventResult)
	}

	si.eventIndicies[spaceID][eventID] = &EventIndex{
		EventID: eventID,
		SpaceID: spaceID,
		PageNum: pageNum,
		Length:  0,
		Hash:    execution.Hash(),
	}

	si.events[spaceID][eventID] = execution
	c, ok := si.eventCounts[spaceID]
	if !ok {
		c = 0
	}

	si.eventCounts[spaceID] = c + 1
	si.eventSpaceIDMappingCache[eventID] = spaceID
	si.eventCache[eventID] = execution
}

func (si *SpaceIndexer) ScanPage(page *model.Page) error {
	log.Printf("indexer scan page %s:%d, tx: %d, ex: %d", page.SpaceID, page.N, len(page.Transactions), len(page.Executions))

	pageNum := page.GetPageNum()
	defer si.SetLastIndexedPageNum(page.SpaceID, pageNum)

	for _, execution := range page.Executions {
		si.AddExecution(pageNum, execution)
	}

	for _, transaction := range page.Transactions {
		err := si.IndexTransaction(transaction)
		if err != nil {
			log.Fatalf("%w", err)
		}
	}

	return nil
}

func (si *SpaceIndexer) IndexMachineStateLogTransaction(tx *model.Transaction) error {
	spaceID, machineID, state := vm.GetMachineStateFromTransaction(tx)

	history := si.GetMachineStateHistory(spaceID, machineID)
	history.AddMachineStateLogTransaction(tx)

	k := si.FormatKeyStateLogTransaction(tx)
	buffer, err := tx.Encode()

	if err != nil {
		return err
	}

	err = si.WriteIndexerDB(k, buffer)
	if err != nil {
		return err
	}

	sk := FormatStateLogTransactionKey(spaceID, machineID, tx.Timestamp)
	return si.WriteIndexerDB(sk, state.Encode())
}

func (si *SpaceIndexer) IndexTransaction(tx *model.Transaction) error {
	log.Printf("index transaction type: %s,", tx.Type)

	switch tx.Type {
	case "log_machine_state":
		return si.IndexMachineStateLogTransaction(tx)
	default:
		break
	}

	return nil
}

func (si *SpaceIndexer) ListEvents(spaceID string, count int) ([]*model.EventResult, error) {
	spaceEvents, ok := si.events[spaceID]
	if !ok {
		return nil, fmt.Errorf("space %s events cache map does not exists", spaceID)
	}

	events := make([]*model.EventResult, 0, 100)
	c := 0

	for _, event := range(spaceEvents) {
		events = append(events, event)

		if c > count {
			return events, nil
		}
		c = c + 1
	}
	
	return events, nil
}

func (si *SpaceIndexer) GetEventCounts() map[string]int64 {
	return si.eventCounts
}

