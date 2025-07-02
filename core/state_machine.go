package core

type StateMachine struct {
	SpaceID			string
	StorageDescriptor 	interface{}

	StateKernel 	StateKernel
	HashKernel 	HashKernel
}

func NewBasicStateMachine(SpaceID string) *StateMachine {
	machine := &StateMachine{
		SpaceID: SpaceID,
		StateKernel: &BasicStateKernel{},
		HashKernel: &BasicHashKernel{},
	}
	return machine
}

func (machine *StateMachine) PreExecuteTransaction(state []float64, tx interface{}) (interface{}, error) {
	return nil, nil
}

func (machine *StateMachine) GetNodeState(size int64) ([]float64, string) {
	stateSeed, payload := machine.StateKernel.GenNodeStateSeed()
	return machine.StateKernel.GenVector(stateSeed, size), payload
}


