package main

import (
	"context"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"goji.io"
	"goji.io/pat"
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

	mux := goji.NewMux()

	req, err := http.NewRequest("POST", "/tx/"+tester.From.Hex()+"/10000000000/2000000/1/a", nil)
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc(pat.Post("/tx/:to/:amount/:gasLimit/:gasPrice/:data"), txHandler(
		ctx,
		client,
		owner.From,
		ownerKey,
	))

	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	client.Commit()

	t.Run("Can proxy a simple tx", func(t *testing.T) {
		want := big.NewInt(10000000000)
		if got, _ := client.BalanceAt(ctx, tester.From, nil); !reflect.DeepEqual(got, want) {
			t.Errorf("txHandler() = %v, want %v", got, want)
		}
	})
}
