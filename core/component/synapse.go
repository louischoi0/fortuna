package component

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/core/vm"
	"fortuna/swift"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const BASIC_NODE_STATE_COUNT = 256

type Synapse struct {
	ID        string
	SpaceID   string
	Epoch     int64
	EpochHash string

	machines map[vm.KernelVersion]*vm.StateMachine
	Universe *vm.Universe

	confirms []*model.EventExecutionResult
	swift    *swift.TCPServer
}

func NewSynapse(spaceID string) *Synapse {
	universe := vm.NewUniverse(spaceID)

	return &Synapse{
		SpaceID:  spaceID,
		Universe: universe,
		Epoch:    0,
		machines: make(map[vm.KernelVersion]*vm.StateMachine),
		confirms: make([]*model.EventExecutionResult, 0, 100),
		swift:    swift.NewServer(),
	}
}

func (n *Synapse) InitMachines() error {
	machine := vm.NewBasicStateMachine(n.Universe, n.SpaceID, 256)
	machine.ResetState()

	n.machines[vm.BaseV000] = machine

	return nil
}

func (n *Synapse) LoadMachine(kernel vm.KernelVersion) (*vm.StateMachine, error) {
	machine, ok := n.machines[kernel]

	if !ok {
		return nil, fmt.Errorf("no machine found for %v", kernel)
	}

	return machine, nil
}

func (n *Synapse) Confirm(request *model.Event) (*model.EventExecutionResult, error) {
	machine, err := n.LoadMachine(vm.KernelVersion(request.Spec.KernelVersion))

	if err != nil {
		return nil, err
	}

	er, err := machine.EmitEventResult(request)

	if er == nil {
		return nil, fmt.Errorf("event result is nil")
	}

	if err != nil {
		return nil, err
	}

	return er, nil
}

func (n *Synapse) Commit(eventresult *model.EventExecutionResult) error {
	return nil
}

func (n *Synapse) ResetMachineState(workspaceID string) {

}

func (n *Synapse) Init() {

}

func (n *Synapse) Run(port int) error {
	if err := n.InitMachines(); err != nil {
		log.Fatalf("Failed to init synapse: %v", err.Error())
	}

	if err := n.swift.Start(port); err != nil {
		log.Fatalf("Failed to start server: %v", err.Error())
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	return nil
}
