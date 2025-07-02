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
	}
	machine.SetStateKernel(BasicStateKernel{})
	machine.SetHashKernel(BasicHashKernel{})
	return machine
}

func (machine *StateMachine) PreExecuteTransaction(state []float64, tx interface{}) (interface{}, error) {
	return nil, nil
}

func (machine *StateMachine) SetStateKernel(k StateKernel) {
	machine.StateKernel = k
}

func (machine *StateMachine) SetHashKernel(k HashKernel) {
	machine.HashKernel = k
}


