package vm

import (
	"time"
	"fortuna/core/model"
)

type ResetStateSignal struct {
	SpaceID   	string
	State     	[]int64
	StateHash 	string
	StateSeed 	int64
}

type StateMachine struct {
	SpaceID string

	State      []int64
	StateHash  string
	StateCount int64

	StateKernel 	StateKernel
	KernelVersion  	KernelVersion

	lastStateGeneratedAt time.Time
	machineCreatedAt     time.Time

	reset_state_signal chan *ResetStateSignal
}

func NewBasicStateMachine(SpaceID string, stateCount int64) *StateMachine {
	machine := &StateMachine{
		SpaceID:            SpaceID,
		StateCount:         stateCount,
		State:              make([]int64, stateCount, stateCount),
		StateKernel:        &BasicStateKernel{},
	}
	return machine
}

func (machine *StateMachine) EmitEventResult(event *model.Event) (*model.EventExecutionResult, error) {
	er := EXEC_INTERFACE(machine, event, machine.StateKernel)
	return er, nil
}

func (machine *StateMachine) ExecuteEvent(state []int64, tx interface{}) (interface{}, error) {
	return nil, nil
}

func (machine *StateMachine) GenState(size int64) ([]int64, string) {
	stateSeed, _ := machine.StateKernel.GenStateSeed()
	return machine.StateKernel.GenVector(stateSeed, size), stateSeed
}

func (machine *StateMachine) ResetState() (string, string) {
	state, seed := machine.GenState(machine.StateCount)
	machine.State = state

	//TODO
	hash := ""
	return seed, hash
}

func (machine *StateMachine) VerifyMachineState() error {
	return nil
}
