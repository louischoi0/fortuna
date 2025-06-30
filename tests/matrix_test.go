package test

import (
	"testing"
	"fmt"
	// "gonum.org/v1/gonum/mat"
	. "github.com/louischoi0/fortuna/core"
)


func TestGonumMatrixKernel_NewDense(t *testing.T) {
	kernel := &GonumMatrixKernel{}
	kernel.NewDense(1,2)
	fmt.Println(kernel)
}

