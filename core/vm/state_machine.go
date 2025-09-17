package vm

import (
	"fortuna/core/model"
	"fortuna/util"
	"time"
	"log"
)

type ResetStateSignal struct {
	SpaceID   string
	State     []int64
	StateHash string
	StateSeed string
}

type StateMachine struct {
	ID      string
	SpaceID string

	State      *model.StateVector
	StateSeed  string
	StateHash  string
	StateCount int64

	StateKernel   StateKernel
	KernelVersion KernelVersion

	lastStateGeneratedAt time.Time
	machineCreatedAt     time.Time

	InitStateTransaction	*model.Transaction
}

func NewBasicStateMachine(space *model.Space, stateCount int64) *StateMachine {
	machine := &StateMachine{
		ID: MakeStateMachineID(space.SpaceID),
		SpaceID:     space.SpaceID,
		StateCount:  stateCount,
		State:       model.NewStateVector(make([]int64, stateCount)),
		StateKernel: &BasicStateKernel{},
	}
	return machine
}

func (machine *StateMachine) EmitEventResult(event *model.Event) (*model.EventResult, error) {
	er := EXEC_INTERFACE(machine, event, machine.StateKernel)
	machine.OnEmit(er)

	return er, nil
}

func VerifyEventResult(state *model.StateVector, kernel StateKernel, res *model.EventResult) bool {
	return VERIFY_INTERFACE(state, res, kernel)
}

func (machine *StateMachine) ExecuteEvent(state *model.StateVector, tx interface{}) (interface{}, error) {
	return nil, nil
}

func (machine *StateMachine) GenState(size int64) (*model.StateVector, string) {
	machine.StateSeed, _ = machine.StateKernel.GenStateSeed()
	machine.State = machine.StateKernel.GenVector(machine.StateSeed, size)

	return machine.State, machine.StateSeed
}

func (machine *StateMachine) ResetState() (string, string) {
	state, seed := machine.GenState(machine.StateCount)
	machine.State = state

	hash := ""
	return seed, hash
}

func (machine *StateMachine) VerifyMachineState() bool {
	return machine.StateKernel.VerifyVector(machine.StateSeed, machine.State)
}

const LOG_MACHINE_STATE_TX_TYPE string = "log_machine_state"

func (machine *StateMachine) NewLogMachineStateTransaction() *model.Transaction {
	// params := structure.NewOrderedMap()
	subroutine := NewLogMachineStateSubroutine(machine.SpaceID, machine.ID, machine.State)

	return &model.Transaction{
		Type: LOG_MACHINE_STATE_TX_TYPE,
		SpaceID: machine.SpaceID,
		From:    machine.ID,
		// Params:  params,
		Operations: subroutine.Operations,
		Timestamp: uint64(util.Now()),
	}
}

func GetMachineStateFromTransaction(tx *model.Transaction) (string, string, *model.StateVector) {
	// machine state should be written to the universe when transaction executed,
	// and retrv machine state for the period have to look up universe db (rocksdb)
	// but now we just extract state from transaction not via universe yet
	if tx.Type != "log_machine_state" {
		log.Fatalf("transaction expected to have log_machine_state type, not %s", tx.Type)
	}
	
	if len(tx.Operations) != 1 {
		log.Fatalf("state log transaction should have one operation. not %d", len(tx.Operations))
	}

	if len(tx.Operations[0].Args) != 3 {
		log.Fatalf("state log transaction should have 2 args. not %d", len(tx.Operations[0].Args))
	}
	
	spaceID := tx.Operations[0].Args[0]
	spaceID, ok := spaceID.(string)
	if !ok {
		log.Fatalf("state log transaction shuld have spaceID for first parameter")
	}
	
	machineID := tx.Operations[0].Args[1]
	machineID, ok = machineID.(string)

	if !ok {
		log.Fatalf("state log transaction shuld have machineID for second parameter")
	}

	v := tx.Operations[0].Args[2]
	sv, ok := v.(*model.StateVector)
	if !ok {
		log.Fatalf("state log transaction shuld have state vector for third parameter not %T", v)
	}

	return spaceID.(string), machineID.(string), sv
}

func MakeStateMachineID(spaceID string) string {
	return util.ConcatHash("state_machine", spaceID)
}


func (m *StateMachine) InitMachineState() *model.Transaction {
	tx := m.NewLogMachineStateTransaction()
	m.InitStateTransaction = tx
	return tx
}


func (machine *StateMachine) OnEmit(res *model.EventResult) {
	if machine.InitStateTransaction == nil {
		log.Fatalf("machine state was not initilized")
	}

	res.RefMachineStateTimestamp = machine.InitStateTransaction.Timestamp
	res.RefMachineID = machine.ID
}

