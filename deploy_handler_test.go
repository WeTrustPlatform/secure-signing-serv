package main

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/WeTrustPlatform/secure-signing-serv/testdata/helloworld"
)

func Test_deployHandler(t *testing.T) {
	ctx := context.Background()

	ownerKey, _ := crypto.GenerateKey()
	owner := bind.NewKeyedTransactor(ownerKey)

	rules := "function validate(tx) return true end"

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

		h := deployHandler(client, rules, owner.From, ownerKey)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		client.Commit()

		if rr.Code != 200 {
			t.Errorf("response code = %v, want %v", rr.Code, 200)
		}

		receipt, err := client.TransactionReceipt(ctx, common.HexToHash(rr.Body.String()))
		if err != nil {
			t.Fatal(err)
		}

		codeAtAddress, err := client.CodeAt(ctx, receipt.ContractAddress, nil)
		if err != nil {
			t.Fatal(err)
		}

		if !strings.Contains(string(byteCode), string(codeAtAddress)) {
			t.Errorf("code at address = %v, want %v", codeAtAddress, byteCode)
		}
	})
}
