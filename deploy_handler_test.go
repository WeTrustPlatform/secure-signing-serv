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
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/WeTrustPlatform/secure-signing-serv/testdata/helloworld"
)

func Test_deployHandler(t *testing.T) {
	ctx := context.Background()

	ownerKey, _ := crypto.GenerateKey()
	owner := bind.NewKeyedTransactor(ownerKey)

	rules := "function validate(tx) return true end"
	signer := types.HomesteadSigner{}

	t.Run("Can deploy a contract", func(t *testing.T) {
		client := backends.NewSimulatedBackend(core.GenesisAlloc{
			owner.From: core.GenesisAccount{Balance: big.NewInt(50000000000)},
		}, 4000000)

		byteCode := common.Hex2Bytes(helloworld.HelloWorldBin[2:])

		query := fmt.Sprintf("/deploy?gasPrice=%d", 1)
		req, err := http.NewRequest("POST", query, bytes.NewBuffer(byteCode))
		if err != nil {
			t.Fatal(err)
			return
		}

		h := deployHandler(client, signer, rules, owner.From, ownerKey)

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		client.Commit()

		if rr.Code != 200 {
			t.Errorf("response code = %v, want %v", rr.Code, 200)
			return
		}

		receipt, err := client.TransactionReceipt(ctx, common.HexToHash(rr.Body.String()))
		if err != nil {
			t.Fatal(err)
			return
		}

		codeAtAddress, err := client.CodeAt(ctx, receipt.ContractAddress, nil)
		if err != nil {
			t.Fatal(err)
			return
		}

		if !strings.Contains(string(byteCode), string(codeAtAddress)) {
			t.Errorf("code at address = %v, want %v", codeAtAddress, byteCode)
			return
		}
	})
}
