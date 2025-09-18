package vm

import (
	"fmt"
	"fortuna/core/model"
	"fortuna/util"
	"log"
)

type EventResultCode string

const (
	SUCCESS EventResultCode = "success"
	FAILED  EventResultCode = "failed"
)

type InterfaceID string

const (
	FIC__001 InterfaceID = "FIC-00-001"
)

func NewErrorEventResultRequiredParameter(event *model.Event, keyName string) *model.EventResult {
	return &model.EventResult{
		Err: &model.EventExecutionError{
			Code:    1,
			Message: fmt.Sprintf("parameter %v required", keyName),
		},
	}
}

func NewErrorEventResultUnknownInterfaceID(event *model.Event) *model.EventResult {
	return &model.EventResult{
		Err: &model.EventExecutionError{
			Code:    1,
			Message: "Unknown Interface ID",
		},
	}
}

func VERIFY_INTERFACE(state *model.StateVector, ex *model.EventResult, kernel StateKernel) bool  {
	switch InterfaceID(ex.Event.Spec.InterfaceID) {
	case FIC__001:
		res := EVENT__001(state, ex.Event, kernel)
		res.RefMachineID = ex.RefMachineID
		res.RefMachineStateTimestamp = ex.RefMachineStateTimestamp

		return ex.Equal(res)
	default:
		return false
	}
}

func EXEC_INTERFACE(machine *StateMachine, event *model.Event, kernel StateKernel) (*model.EventResult, uint64) {
	var res *model.EventResult
	salt := machine.GetSalt()

	switch InterfaceID(event.Spec.InterfaceID) {
	case FIC__001:
		res = EVENT__001(machine.State, event, kernel)
		return res, salt
	default:
		return NewErrorEventResultUnknownInterfaceID(event), 0
	}

}

func EVENT__001(machineState *model.StateVector, event *model.Event, kernel StateKernel) *model.EventResult {
	seed := event.Hash()
	slot_count, ok := event.Spec.Params.UInt64("slot_count")

	if !ok {
		return NewErrorEventResultRequiredParameter(event, "slot_count")
	}

	if int(slot_count) >= machineState.Size() {
		return NewErrorEventResultRequiredParameter(event, "slot_count is bigger than legth of machine state")
	}

	indices := kernel.GenIndex(seed, 0, slot_count, 2)
	a := indices[0] % uint64(slot_count)
	b := indices[1] % uint64(slot_count)

	log.Printf("seed=%v, a=%v, b=%v", seed, a, b)
	log.Printf("sa=%v, sb=%v", machineState.Get(int64(a)), machineState.Get(int64(b)))

	var result string
	if machineState.Get(int64(a)) > machineState.Get(int64(b)) {
		result = "t"
	} else {
		result = "f"
	}

	executionResult := model.NewEventResultFromEvent(event, result)
	return executionResult
}


func EVENT__002(machineState *model.StateVector, event *model.Event, kernel StateKernel) *model.EventResult {
	seed := event.Hash()
	subset_count, ok := event.Spec.Params.UInt64("subset_count")

	if !ok {
		return NewErrorEventResultRequiredParameter(event, "subset_count")
	}

	if int(subset_count) >= machineState.Size() {
		return NewErrorEventResultRequiredParameter(event, "subset_count is bigger than legth of machine state")
	}

	indices := kernel.GenIndexU(seed, 0, subset_count, 2)

	var result string
	vidx := util.IndexOfMax(indices)

	if vidx == 0 {
		result = "t"
	} else {
		result = "f"
	}

	executionResult := model.NewEventResultFromEvent(event, result)
	return executionResult
}

