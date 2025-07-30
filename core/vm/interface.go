package vm

import (
	"fortuna/core/model"
	"fmt"
)

type EventResultCode string

const (
	SUCCESS EventResultCode = "success"
	FAILED  EventResultCode = "failed"
)

type InterfaceID string

const (
	FIC__001 InterfaceID = "FIC_001"
)

func NewErrorEventExecutionResultRequiredParameter(event *model.Event, keyName string) *model.EventExecutionResult {
	return &model.EventExecutionResult{
		Err:    &model.EventExecutionError{
			Code: 1, 
			Message: fmt.Sprintf("parameter %v required", keyName),
		},
	}
}


func NewErrorEventExecutionResultUnknownInterfaceID(event *model.Event) *model.EventExecutionResult {
	return &model.EventExecutionResult{
		Err:    &model.EventExecutionError{
			Code: 1, 
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
	seed, _ := kernel.GetEventHash(event)
	slotCount, ok := event.Spec.Params.Int64("slot_count")

	if !ok {
		return NewErrorEventExecutionResultRequiredParameter(event, "slot_count")
	}

	indices := kernel.GenIndex(seed, 2, 0, slotCount)

	var result string
	if machine.State[indices[0]] > machine.State[indices[1]] {
		result = "t"
	} else {
		result = "f"
	}

	executionResult := &model.EventExecutionResult{
		Result: result,
		EventHash:   event.Hash(),
		Err:    nil,
	}

	return executionResult
}
