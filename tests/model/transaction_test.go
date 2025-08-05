package test

import (
	"fortuna/core/model"
	"fortuna/structure"
	"fmt"
	"testing"
)

func TestTransaction(t *testing.T) {
	t.Run("test transaction hash", func(t *testing.T) {
		
		transaction := model.NewTransaction("space_id", "from", structure.NewOrderedMap())
		transaction.Timestamp = 1754354451836053510

		transaction.AddOperation(model.NewOperation("x0", "write_var", []interface{}{1}))
		transaction.UpdateRawCode()

		hash := transaction.Hash()
		fmt.Println(hash)
	})
}
