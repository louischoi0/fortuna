package vm

import (
	"time"
)

type ResetStateSignal struct {
	SpaceID		string
	State		[]float64
	StateHash	string
	StateSeed	int64
}

type StateMachine struct {
	SpaceID			string

	State 		[]float64
	StateHash	string	
	StateCount	int64

	StateKernel 	StateKernel
	
	lastStateGeneratedAt	time.Time
	machineCreatedAt	time.Time

	reset_state_signal chan *ResetStateSignal
}

func NewBasicStateMachine(SpaceID string, stateCount int64, reset_chann chan *ResetStateSignal) *StateMachine {
	machine := &StateMachine{
		SpaceID: SpaceID,
		StateCount:	stateCount,
		State: make([]float64, stateCount, stateCount),
		StateKernel: &BasicStateKernel{},
		reset_state_signal: reset_chann,
	}
	return machine
}

func (machine *StateMachine) EmitEventResult(event model.Event) (*model.EventExecutionResult, error) {

}

func (machine *StateMachine) ExecuteTransaction(state []float64, tx interface{}) (interface{}, error) {
	return nil, nil
}

func (machine *StateMachine) GenNodeState(size int64) ([]float64, int64) {
	stateSeed, _ := machine.StateKernel.GenStateSeed()
	return machine.StateKernel.GenVector(stateSeed, size), stateSeed
}

func (machine *StateMachine) ResetState() (int64, string) {
	state, seed := machine.GenNodeState(machine.StateCount)
	machine.State = state
	
	hash := machine.StateKernel.HashState(state)

	return seed, hash
}

func (machine *StateMachine) VerifyMachineState() error {
}



