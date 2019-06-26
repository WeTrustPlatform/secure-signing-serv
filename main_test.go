package main

import (
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
	"github.com/ethereum/go-ethereum/crypto"
)

func Test_txHandler(t *testing.T) {
	ctx := context.Background()

	ownerKey, _ := crypto.GenerateKey()
	owner := bind.NewKeyedTransactor(ownerKey)

	testerKey, _ := crypto.GenerateKey()
	tester := bind.NewKeyedTransactor(testerKey)

	client := backends.NewSimulatedBackend(core.GenesisAlloc{
		owner.From: core.GenesisAccount{Balance: big.NewInt(50000000000)},
	}, 4000000)

	var amount int64 = 10000000000
	query := fmt.Sprintf("/tx?to=%s&amount=%d&gasLimit=2000000&gasPrice=1", tester.From.Hex(), amount)
	req, err := http.NewRequest("POST", query, nil)
	if err != nil {
		t.Fatal(err)
	}

	h := txHandler(ctx, client, owner.From, ownerKey)

	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	client.Commit()

	t.Run("Can proxy a simple tx", func(t *testing.T) {
		want := big.NewInt(amount)
		if got, _ := client.BalanceAt(ctx, tester.From, nil); !reflect.DeepEqual(got, want) {
			t.Errorf("txHandler() = %v, want %v", got, want)
		}
	})
}
