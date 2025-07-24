package vm

import (
	"fortuna/structure"
)

type EventResultCode string

const (
	SUCCESS EventResultCode = "success"
	FAILED  EventResultCode = "failed"
)

func SPEC__001(kernelVersion KernelVersion, params *structure.OrderedMap) *EventSpec {
	return &EventSpec{
		Protocol: EventInterface{
			InterfaceID:   1,
			InterfaceName: "FIC_001",
		},
		Kernel: kernelVersion,
		Params: params,
	}
}

func EVENT__001(spec *EventSpec) *EventExecutionResult {
	/**
	kernel := LoadKernel(spec.Kernel)
	seed, _ := kernel.GetEventSeed(spec)

	slotCount := spec.Param("slot_count").AsInt64()
	indices := kernel.GenIndex(seed, 0, slotCount, 2)
	**/

	executionResult := &EventExecutionResult{
		Result: "success",
		Err:    nil,
	}

	return executionResult

}
