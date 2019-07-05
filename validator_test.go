package main

import (
	"math/big"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
)

func Test_validate(t *testing.T) {
	t.Run("Can validate a transaction", func(t *testing.T) {
		to := common.HexToAddress("0x5597285BbE81BaF351e2C0884e9a5f4416958862")
		call := ethereum.CallMsg{
			To:    &to,
			Value: big.NewInt(10000000000),
			Data:  []byte("abcd"),
		}

		rules := `
function validate(tx)
	return tx.to == "0x5597285BbE81BaF351e2C0884e9a5f4416958862" or tx.value == "10000000000"
end
`
		got, _ := validate(rules, call)
		want := true
		if !reflect.DeepEqual(got, want) {
			t.Errorf("validate = %v, want %v", got, want)
		}
	})

	t.Run("Can invalidate a transaction", func(t *testing.T) {
		to := common.HexToAddress("0x5597285BbE81BaF351e2C0884e9a5f4416958861")
		call := ethereum.CallMsg{
			To:    &to,
			Value: big.NewInt(10000000000),
			Data:  []byte("abcd"),
		}

		rules := `
function validate(tx)
	return tx.to == "0x5597285BbE81BaF351e2C0884e9a5f4416958862" or tx.value == "10000000000"
end
`
		got, _ := validate(rules, call)
		want := false
		if !reflect.DeepEqual(got, want) {
			t.Errorf("validate = %v, want %v", got, want)
		}
	})

	t.Run("Can filter on data field", func(t *testing.T) {
		to := common.HexToAddress("0x5597285BbE81BaF351e2C0884e9a5f4416958861")
		call := ethereum.CallMsg{
			To:    &to,
			Value: big.NewInt(10000000000),
			Data:  []byte("Hello world!"),
		}

		rules := `
function validate(tx)
	return tx.data == "Hello world!"
end
`
		got, _ := validate(rules, call)
		want := true
		if !reflect.DeepEqual(got, want) {
			t.Errorf("validate = %v, want %v", got, want)
		}
	})

}
