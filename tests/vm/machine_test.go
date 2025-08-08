package test

import (
	"fmt"
	. "fortuna/core/vm"
	"testing"
)

func TestMachine(t *testing.T) {
	machine := NewBasicStateMachine(nil, "thisisspace", 256)
	seed, _ := machine.ResetState()

	res := machine.StateKernel.VerifyVector(seed, machine.State)
	fmt.Println(res)
	res = machine.StateKernel.VerifyVector("abcdef", machine.State)
	fmt.Println(res)

	tx := machine.NewLogMachineStateTransaction()
	encoded, err := tx.Encode()
	if err != nil {
		t.Fatalf("failed to encode transaction: %v", err)
	}

	fmt.Println(encoded)
}
