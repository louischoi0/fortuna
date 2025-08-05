package test

import (
        . "fortuna/core/vm"
        "testing"
	"fmt"
)

func TestMachine(t *testing.T) {
        machine := NewBasicStateMachine(nil, "thisisspace", 256)
        seed, _ := machine.ResetState()

	res := machine.StateKernel.VerifyVector(seed, machine.State)
        fmt.Println(res)                                                        
	res = machine.StateKernel.VerifyVector("abcdef", machine.State)
        fmt.Println(res)                                                        
}                                                                               
