package test

import (
	"testing"
	"fmt"
	. "fortuna/core"
)


func TestGonumMatrixKernel_NewDense(t *testing.T) {
	kernel := &GonumMatrixKernel{}
	kernel.NewDense(1,2)
	fmt.Println(kernel)
}

