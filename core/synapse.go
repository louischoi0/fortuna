package core

import (
	"fortuna/swift"
	"fortuna/core/vm"
	"fortuna/core/model"
	"log"
	"os"
        "os/signal"
        "syscall"
)

const BASIC_NODE_STATE_COUNT = 256

type Synapse struct {
	Epoch		int64
	EpochHash	string	

	machines 	map[string]*vm.StateMachine

	confirms 	[]*model.EventExecutionResult
	swift *swift.TCPServer
}

func NewBasicSynapse(spaceID string) *Synapse {
	return &Synapse{
		Epoch: 0,
		machines: make(map[string]*vm.StateMachine),
		confirms: make([]*model.EventExecutionResult, 0, 100),
		swift: swift.NewServer(),
	}
}

func (n *Synapse) LoadMachine(kernel vm.KernelVersion) (*vm.StateMachine, error) {
	return nil, nil
}

func (n *Synapse) Confirm(request *model.Event) (*model.EventExecutionResult, error) {
	er := model.EventExecutionResult{}
	// machine, err := n.LoadMachine()
	return &er, nil
}

func (n *Synapse) Commit(eventresult *model.EventExecutionResult) error {
	return nil
}

func (n *Synapse) ResetMachineState(workspaceID string) {

}

func (n *Synapse) Init() {

}

func (n *Synapse) Run(port int) error {

	if err := n.swift.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

        <-sigChan
	return nil
}

