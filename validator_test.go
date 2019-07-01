package main

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func Test_validate(t *testing.T) {
	t.Run("can validate a transaction", func(t *testing.T) {
		tx := types.NewTransaction(
			1,
			common.HexToAddress("0x5597285BbE81BaF351e2C0884e9a5f4416958862"),
			big.NewInt(1),
			21000,
			big.NewInt(10000000000),
			[]byte("abcd"),
		)

		rules := `
function validate(tx)
	return tx.to == "0x5597285BbE81BaF351e2C0884e9a5f4416958862" or tx.value == "10000000000"
end
`
		got, _ := validate(rules, tx)
		want := true
		if !reflect.DeepEqual(got, want) {
			t.Errorf("validate = %v, want %v", got, want)
		}
	})

	t.Run("can invalidate a transaction", func(t *testing.T) {
		tx := types.NewTransaction(
			1,
			common.HexToAddress("0x5597285BbE81BaF351e2C0884e9a5f4416958861"),
			big.NewInt(1),
			21000,
			big.NewInt(10000000000),
			[]byte("abcd"),
		)

		rules := `
function validate(tx)
	return tx.to == "0x5597285BbE81BaF351e2C0884e9a5f4416958862" or tx.value == "10000000000"
end
`
		got, _ := validate(rules, tx)
		want := false
		if !reflect.DeepEqual(got, want) {
			t.Errorf("validate = %v, want %v", got, want)
		}
	})

	t.Run("can filter on data field", func(t *testing.T) {
		tx := types.NewTransaction(
			1,
			common.HexToAddress("0x5597285BbE81BaF351e2C0884e9a5f4416958861"),
			big.NewInt(1),
			21000,
			big.NewInt(10000000000),
			[]byte("Hello world!"),
		)

		rules := `
function validate(tx)
	return tx.data == "Hello world!"
end
`
		got, _ := validate(rules, tx)
		want := true
		if !reflect.DeepEqual(got, want) {
			t.Errorf("validate = %v, want %v", got, want)
		}
	})

}
