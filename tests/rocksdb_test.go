package test

import (
	"fmt"
	"fortuna/rock"
	"testing"
)

func TestRocksDB(t *testing.T) {
	t.Run("encode/decode block index", func(t *testing.T) {
		testDB, err := rock.GetDBInstance("test")
		if err != nil {
			t.Fatalf("failed to get test db: %v", err)
		}
		defer rock.CloseDB(testDB)

		key := "hello"
		value := "world"

		err = rock.SetValue(testDB, key, []byte(value))
		if err != nil {
			t.Fatalf("failed to write: %v", err)
		}

		v, err := rock.GetValue(testDB, key)
		if err != nil {
			t.Fatalf("failed to read: %v", err)
		}

		if string(v) != "world" {
			t.Fatalf("invalid value: %s", string(v))
		}
		fmt.Println(string(v))
	})
}
