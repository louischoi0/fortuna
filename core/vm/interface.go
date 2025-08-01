package vm

import (
	"fmt"
	"fortuna/core/model"
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

func NewErrorEventExecutionResultRequiredParameter(event *model.Event, keyName string) *model.EventExecutionResult {
	return &model.EventExecutionResult{
		Err: &model.EventExecutionError{
			Code:    1,
			Message: fmt.Sprintf("parameter %v required", keyName),
		},
	}
}

func NewErrorEventExecutionResultUnknownInterfaceID(event *model.Event) *model.EventExecutionResult {
	return &model.EventExecutionResult{
		Err: &model.EventExecutionError{
			Code:    1,
			Message: "Unknown Interface ID",
		},
	}
}

func EXEC_INTERFACE(machine *StateMachine, event *model.Event, kernel StateKernel) *model.EventExecutionResult {
	switch InterfaceID(event.Spec.InterfaceID) {
	case FIC__001:
		return EVENT__001(machine, event, kernel)
	default:
		return NewErrorEventExecutionResultUnknownInterfaceID(event)
	}

}

func EVENT__001(machine *StateMachine, event *model.Event, kernel StateKernel) *model.EventExecutionResult {
	seed := event.Hash()
	slot_count, ok := event.Spec.Params.Int64("slot_count")

	if !ok {
		return NewErrorEventExecutionResultRequiredParameter(event, "slot_count")
	}

	if int(slot_count) >= len(machine.State) {
		return NewErrorEventExecutionResultRequiredParameter(event, "slot_count is bigger than legth of machine state")
	}

	indices := kernel.GenIndex(seed, 0, slot_count, 2)
	a := indices[0] % slot_count
	b := indices[1] % slot_count

	fmt.Printf("seed=%v, a=%v, b=%v, sa=%v, ba=%v", seed, a, b, machine.State[a], machine.State[b])

	var result string
	if machine.State[a] > machine.State[b] {
		result = "t"
	} else {
		result = "f"
	}

	executionResult := model.NewEventExecutionResultFromEvent(event, result)
	return executionResult
}
