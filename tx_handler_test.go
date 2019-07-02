package main

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func Test_txHandler(t *testing.T) {
	ownerKey, _ := crypto.GenerateKey()
	owner := bind.NewKeyedTransactor(ownerKey)

	testerKey, _ := crypto.GenerateKey()
	tester := bind.NewKeyedTransactor(testerKey)

	rules := `function validate(tx) return true end`
	signer := types.HomesteadSigner{}

	t.Run("Can transact with value", func(t *testing.T) {
		client := backends.NewSimulatedBackend(core.GenesisAlloc{
			owner.From: core.GenesisAccount{Balance: big.NewInt(50000000000)},
		}, 4000000)

		query := fmt.Sprintf("/tx?to=%s&amount=%d&gasPrice=%d", tester.From.Hex(), 10000000000, 1)
		req, err := http.NewRequest("POST", query, bytes.NewBufferString(""))
		if err != nil {
			t.Fatal(err)
			return
		}

		h := txHandler(client, signer, rules, owner.From, ownerKey)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		client.Commit()

		if rr.Code != 200 {
			t.Errorf("response code = %v, want %v", rr.Code, 200)
		}

		want := big.NewInt(10000000000)
		if got, _ := client.BalanceAt(context.Background(), tester.From, nil); !reflect.DeepEqual(got, want) {
			t.Errorf("tester balance = %v, want %v", got, want)
			return
		}
	})

	t.Run("Can transact with data", func(t *testing.T) {
		client := backends.NewSimulatedBackend(core.GenesisAlloc{
			owner.From: core.GenesisAccount{Balance: big.NewInt(50000000000)},
		}, 4000000)

		query := fmt.Sprintf("/tx?to=%s&amount=%d&gasPrice=%d", tester.From.Hex(), 10000000000, 1)
		req, err := http.NewRequest("POST", query, bytes.NewBufferString("abcdef"))
		if err != nil {
			t.Fatal(err)
			return
		}

		h := txHandler(client, signer, rules, owner.From, ownerKey)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		client.Commit()

		want := big.NewInt(10000000000)
		if got, _ := client.BalanceAt(context.Background(), tester.From, nil); !reflect.DeepEqual(got, want) {
			t.Errorf("tester balance = %v, want %v", got, want)
			return
		}
	})
}
