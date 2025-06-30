package core

import (
	"gonum.org/v1/gonum/mat"
)

type MatrixKernel interface {
	NewDense(dims ...int) mat.Vector
}

type GonumMatrixKernel struct {

}

func (g *GonumMatrixKernel) NewDense(dims ...int) mat.Vector {
	if len(dims) != 2 {
		panic("NewDense requires exactly two dimensions: size and unused")
	}
	size := dims[0]
	data := make([]float64, size)

	for i := range data {
		data[i] = 0.0
	}

	return mat.NewVecDense(size, data)
}





