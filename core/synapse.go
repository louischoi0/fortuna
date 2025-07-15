package core

import (
	"fortuna/swift"
	"fortuna/core/vm"
)

const BASIC_NODE_STATE_COUNT = 256

type Synapse struct {
	Epoch		int64
	EpochHash	string	

	// kernel: *state_machine
	machines 	map[string]*vm.StateMachine

	confirms 	[]*model.ExeuctionResult
	swift *swift.TCPServer
}

func NewBasicSynapse(spaceID string) *Synapse {
	return &Synapse{
		Epoch: 0,
		machines: make(map[string]*vm.StateMachine),
		confirms: make([]*model.ExecutionResult, 0, 100),
		swift: swift.NewServer(),
	}
}

func (n *Synapse) Confirm(event *model.Event) (*model.EventResult, error) {
	er := EventResult{}

	return &er, nil
}

func (n *Synapse) Commit(eventresult *model.EventResult) error {
	return nil
}

func (n *Synapse) ResetMachineState(workspaceID string) {

}

func (n *Synapse) Init() {

}
