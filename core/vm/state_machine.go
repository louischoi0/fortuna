package vm

import (
	"fortuna/core/model"
	"time"
)

type ResetStateSignal struct {
	SpaceID   string
	State     []int64
	StateHash string
	StateSeed string
}

type StateMachine struct {
	ID         string
	SpaceID    string
	Universe   *Universe
	LastHeight int64

	State      *model.StateVector
	StateSeed  string
	StateHash  string
	StateCount int64

	StateKernel   StateKernel
	KernelVersion KernelVersion

	lastStateGeneratedAt time.Time
	machineCreatedAt     time.Time

	reset_state_signal chan *ResetStateSignal
}

func NewBasicStateMachine(universe *Universe, SpaceID string, stateCount int64) *StateMachine {
	machine := &StateMachine{
		Universe:    universe,
		SpaceID:     SpaceID,
		StateCount:  stateCount,
		State:       model.NewStateVector(make([]int64, stateCount)),
		StateKernel: &BasicStateKernel{},
	}
	return machine
}

func (machine *StateMachine) EmitEventResult(event *model.Event) (*model.EventResult, error) {
	er := EXEC_INTERFACE(machine, event, machine.StateKernel)
	return er, nil
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

func (machine *StateMachine) NewLogMachineStateTransaction() *model.Transaction {
	// params := structure.NewOrderedMap()
	subroutine := NewLogMachineStateSubroutine(machine.ID, machine.State)

	return &model.Transaction{
		SpaceID: machine.SpaceID,
		From:    machine.ID,
		// Params:  params,
		Operations: subroutine.Operations,
	}
}
