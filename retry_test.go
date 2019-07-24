package main

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/WeTrustPlatform/secure-signing-serv/sss"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func Test_retry(t *testing.T) {
	ownerKey, _ := crypto.GenerateKey()
	owner := bind.NewKeyedTransactor(ownerKey)

	testerKey, _ := crypto.GenerateKey()
	tester := bind.NewKeyedTransactor(testerKey)

	rules := `function validate(tx) return true end`
	signer := types.HomesteadSigner{}

	db := &dbMock{}

	t.Run("Can replace a transaction", func(t *testing.T) {
		client := backends.NewSimulatedBackend(core.GenesisAlloc{
			owner.From: core.GenesisAccount{Balance: big.NewInt(50000000000)},
		}, 4000000)

		p := sss.TxPayload{To: tester.From.Hex(), Value: "10000000000", GasPrice: "1"}
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(p)
		req, err := http.NewRequest("POST", "/v1/proxy/transactions", b)
		if err != nil {
			t.Fatal(err)
			return
		}

		h := txHandler(client, signer, rules, owner.From, ownerKey, db)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		if rr.Code != 200 {
			t.Errorf("response code = %v, want %v", rr.Code, 200)
		}
		client.Rollback()

		want := big.NewInt(0)
		if got, _ := client.BalanceAt(context.Background(), tester.From, nil); !reflect.DeepEqual(got, want) {
			t.Errorf("tester balance = %v, want %v", got, want)
			return
		}

		// Retry

		hash := rr.Body.String()
		rp := sss.RetryPayload{Hash: hash, GasPrice: "2"}
		rb := new(bytes.Buffer)
		json.NewEncoder(rb).Encode(rp)
		rreq, err := http.NewRequest("POST", "/v1/proxy/transactions/retry", rb)
		if err != nil {
			t.Fatal(err)
			return
		}

		rh := retryHandler(client, signer, rules, owner.From, ownerKey, db)
		rrr := httptest.NewRecorder()
		rh.ServeHTTP(rrr, rreq)
		client.Commit()

		if rrr.Code != 200 {
			t.Errorf("response code = %v, want %v", rrr.Code, 200)
		}

		want = big.NewInt(10000000000)
		if got, _ := client.BalanceAt(context.Background(), tester.From, nil); !reflect.DeepEqual(got, want) {
			t.Errorf("tester balance = %v, want %v", got, want)
			return
		}
	})
}
