package main

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
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
	ownerAuth := bind.NewKeyedTransactor(ownerKey)

	testerKey, _ := crypto.GenerateKey()
	tester := bind.NewKeyedTransactor(testerKey)

	client := backends.NewSimulatedBackend(core.GenesisAlloc{
		owner.From: core.GenesisAccount{Balance: big.NewInt(50000000000)},
	}, 4712388)

	mux := goji.NewMux()

	req, err := http.NewRequest("POST", "/tx/"+tester.From.Hex()+"/10000000000/2000000/1/a", nil)
	if err != nil {
		t.Fatal(err)
	}

	mux.HandleFunc(pat.Post("/tx/:to/:amount/:gasLimit/:gasPrice/:data"), txHandler(ctx, client, ownerAuth, ownerKey))

	mux.ServeHTTP(httptest.NewRecorder(), req)

	client.Commit()

	t.Run("Can proxy a simple tx", func(t *testing.T) {
		// if got := txHandler(tt.args.ctx, tt.args.client, tt.args.key); !reflect.DeepEqual(got, tt.want) {
		// 	t.Errorf("txHandler() = %v, want %v", got, tt.want)
		// }
		fmt.Println(client.BalanceAt(ctx, tester.From, nil))
	})
}
