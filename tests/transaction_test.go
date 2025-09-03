package test

import (
	//"bytes"
	"fmt"
	"log"
	"fortuna/core/model"
	//"encoding/json"
	"fortuna/core/vm"
	"testing"
)

func TestTransaction(t *testing.T) {
	t.Run("test transaction hash", func(t *testing.T) {

		v := model.NewStateVector(make([]int64, 8, 8))

		transaction := model.NewRawTransaction("space_id", "from")
		transaction.SpaceID = "0000000000000000000000000000000000000000000000000000000000000000"
		transaction.From = "0000000000000000000000000000000000000000000000000000000000000000"
		transaction.Timestamp = 1754354451836053510
		abi := vm.ABI{}
		op := abi.WriteVar("k", v.Var())

		transaction.AddOperation(op)
		transaction.UpdateRawCode()

		hash := transaction.Hash()
		fmt.Println(hash)

		encoded, err := transaction.Encode()
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		decoded, err := model.DecodeRawTransaction(encoded)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}
	
		va := decoded.Operations[0].Args[1]
		//log.Println(va)
		vv, _ := va.(*model.Var)
		log.Printf("%T", vv.Vec())
	})
}


/**
func aTestMachineStateTransaction(t *testing.T) {
	machine := &vm.StateMachine{
		ID: "0000000000000000000000000000000000000000000000000000000000000000",
		SpaceID: "0000000000000000000000000000000000000000000000000000000000000000",
		StateCount:  10,
		State:       model.NewStateVector(make([]int64, 10)),
		StateKernel: &vm.BasicStateKernel{},
	}

	st := machine.NewLogMachineStateTransaction()
	buf, err := st.Encode()
	if err != nil {
		log.Println(err.Error())
	}

	tx, err := model.DecodeRawTransaction(buf)
	if err != nil {
		log.Println(err.Error())
	}

	b, err := json.Marshal(tx)
	if err != nil {
		log.Println(err.Error())
	}

	log.Println(b)

	v := tx.Operations[0].Args[1]
	_, ok := v.(model.Var)
	if !ok {
		log.Printf("conversion failed %t", v)
		log.Println(v)
	}
}
**/
/**
func TestTransaction(t *testing.T) {
	t.Run("test transaction hash", func(t *testing.T) {

		transaction := model.NewRawTransaction("space_id", "from")
		transaction.Timestamp = 1754354451836053510

		transaction.AddOperation(model.NewOperation("x0", "write_var", []interface{}{1}))
		transaction.UpdateRawCode()

		hash := transaction.Hash()
		fmt.Println(hash)
	})
}

func TestRawTransaction(t *testing.T) {
	t.Run("encode/decode and verify hash", func(t *testing.T) {
		// Given: Create a raw transaction
		rtx := model.NewRawTransaction("space_id___________________________________________1234560000000", "from_address________________________________________567800000000")
		rtx.Timestamp = 1754354451836053510

		//write_op := abi.WriteVar(machineID, stateVector.Var())
		//subroutine.AddOperation(write_op)

		rtx.AddOperation(model.NewOperation("x0", "write_var", []interface{}{2}))

		// When: encode the transaction
		encoded, err := rtx.Encode()
		if err != nil {
			t.Fatalf("failed to encode transaction: %v", err)
		}

		// And: decode the transaction
		decoded, err := model.DecodeRawTransaction(encoded)
		if err != nil {
			t.Fatalf("failed to decode transaction: %v", err)
		}

		// Then: verify original and decoded transaction match
		if rtx.Timestamp != decoded.Timestamp {
			t.Errorf("timestamp mismatch: got %d, want %d", decoded.Timestamp, rtx.Timestamp)
		}
		if rtx.SpaceID != decoded.SpaceID {
			t.Errorf("space_id mismatch: got %s, want %s", decoded.SpaceID, rtx.SpaceID)
		}
		if rtx.From != decoded.From {
			t.Errorf("from mismatch: got %s, want %s", decoded.From, rtx.From)
		}
		if !bytes.Equal(rtx.RawCode, decoded.RawCode) {
			t.Errorf("raw_code mismatch: got %v, want %v", decoded.RawCode, rtx.RawCode)
		}
		if len(decoded.Operations) != 1 {
			t.Errorf("expected 1 operation, got %d", len(decoded.Operations))
		}
		if decoded.Operations[0].OpCode != "x0" {
			t.Errorf("operation object mismatch: got %s", decoded.Operations[0].OpCode)
		}

		// Finally: hash match
		if rtx.Hash() != decoded.Hash() {
			t.Errorf("hash mismatch: encoded=%s, decoded=%s", rtx.Hash(), decoded.Hash())
		}

		fmt.Println(rtx.Hash())
		fmt.Println(decoded.Hash())
		fmt.Println(string(rtx.RawCode))
		fmt.Println(string(decoded.RawCode))
		fmt.Println(decoded.Operations)
		fmt.Println(rtx.Operations)
		//fmt.Println(string(encoded))
	})
}
**/
