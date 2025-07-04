package vm


import (
	"fortuna/core/model"
)

type Interpreter struct {
	kernel		StateKernel

	event		*model.Event

	result		string
	errMsg		string
	success		bool
}


func (i *Interpreter) SetKernel(k StateKernel) {
	i.kernel = k
}
