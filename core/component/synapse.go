package component

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/core/vm"
	"fortuna/util"
	"log"
)

const BASIC_NODE_STATE_COUNT = 256

type Synapse struct {
	ID        string
	Space     *model.Space
	Epoch     int64
	EpochHash string

	machines map[vm.KernelVersion]*vm.StateMachine
	confirms []*model.EventResult
}

func NewSynapse(space *model.Space) *Synapse {

	return &Synapse{
		ID: MakeSynapseID(space.SpaceID),
		Space:    space,
		Epoch:    0,
		machines: make(map[vm.KernelVersion]*vm.StateMachine),
		confirms: make([]*model.EventResult, 0, 100),
	}
}

func (n *Synapse) InitMachines() ([]*model.Transaction, error) {
	initMachineTxs := make([]*model.Transaction, 0, 8)
	
	machine := vm.NewBasicStateMachine(n.Space, 256)
	machine.ResetState()

	n.machines[vm.BaseV000] = machine
	tx := machine.InitMachineState()

	initMachineTxs = append(initMachineTxs, tx)

	return initMachineTxs, nil
}

func (n *Synapse) LoadMachine(kernel vm.KernelVersion) (*vm.StateMachine, error) {
	machine, ok := n.machines[kernel]

	if !ok {
		return nil, fmt.Errorf("no machine found for %v", kernel)
	}

	return machine, nil
}

func (n *Synapse) Confirm(request *model.Event) (*model.EventResult, error) {
	machine, err := n.LoadMachine(vm.KernelVersion(request.Spec.KernelVersion))
	log.Printf("machine %s loaded", machine.ID)

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

func (n *Synapse) Commit(eventresult *model.EventResult) error {
	return nil
}

func (n *Synapse) ResetMachineState(workspaceID string) {

}

func (n *Synapse) Init() {

}

func (n *Synapse) Bootstrap() ([]*model.Transaction, error) {
	txs, err := n.InitMachines()

	if err != nil {
		log.Fatalf("Failed to init synapse: %v", err.Error())
	}

	return txs, nil
}

func MakeSynapseID(spaceID string) string {
	return util.ConcatHash("synapse", spaceID)
}


