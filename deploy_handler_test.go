package main

import (
	"bytes"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/common"

	"github.com/WeTrustPlatform/secure-signing-serv/testdata/helloworld"
)

func Test_deployHandler(t *testing.T) {
	ownerKey, _ := crypto.GenerateKey()
	owner := bind.NewKeyedTransactor(ownerKey)

	t.Run("Can proxy a simple contract deployment", func(t *testing.T) {
		client := backends.NewSimulatedBackend(core.GenesisAlloc{
			owner.From: core.GenesisAccount{Balance: big.NewInt(50000000000)},
		}, 4000000)

		byteCode := common.Hex2Bytes(helloworld.HelloWorldBin[2:])

		query := fmt.Sprintf("/deploy?gasPrice=%d", 1)
		req, err := http.NewRequest("POST", query, bytes.NewBuffer(byteCode))
		if err != nil {
			t.Fatal(err)
		}

		h := deployHandler(client, owner.From, ownerKey)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		client.Commit()

		fmt.Println(rr.Body)

		// want := big.NewInt(10000000000)
		// if got, _ := client.BalanceAt(context.Background(), tester.From, nil); !reflect.DeepEqual(got, want) {
		// 	t.Errorf("tester balance = %v, want %v", got, want)
		// }
	})
}
