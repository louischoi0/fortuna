package core

type StateMachine struct {
	SpaceID			string
	StorageDescriptor 	interface{}

	StateKernel 	*StateKernel
	HashKernel 	*HashKernel
}

func (machine *StateMachine) PreExecuteTransaction(state []float64, tx interface{}) (interface{}, error) {
	return nil, nil
}

func (machine *StateMachine) SetStateKernel(k *StateKernel) {
	machine.StateKernel = k
}

func (machine *HashKernel) SetHashKernel(k *HashKernel) {
	machine.HashKernel = k
}

