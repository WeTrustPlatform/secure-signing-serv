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

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/WeTrustPlatform/secure-signing-serv/testdata/helloworld"
)

func Test_methodCall(t *testing.T) {
	ctx := context.Background()

	ownerKey, _ := crypto.GenerateKey()
	owner := bind.NewKeyedTransactor(ownerKey)

	rules := "function validate(tx) return true end"
	signer := types.HomesteadSigner{}

	t.Run("Can call a method of a smart contract", func(t *testing.T) {
		client := backends.NewSimulatedBackend(core.GenesisAlloc{
			owner.From: core.GenesisAccount{Balance: big.NewInt(50000000000)},
		}, 4000000)

		byteCode := common.Hex2Bytes(helloworld.HelloWorldBin[2:])

		deployQuery := fmt.Sprintf("/deploy?gasPrice=%d", 1)
		req, err := http.NewRequest("POST", deployQuery, bytes.NewBuffer(byteCode))
		if err != nil {
			t.Fatal(err)
		}

		h := deployHandler(client, signer, rules, owner.From, ownerKey)

		deployRR := httptest.NewRecorder()
		h.ServeHTTP(deployRR, req)

		client.Commit()

		if deployRR.Code != 200 {
			t.Errorf("response code = %v, want %v", deployRR.Code, 200)
		}

		deployReceipt, err := client.TransactionReceipt(ctx, common.HexToHash(deployRR.Body.String()))
		if err != nil {
			t.Fatal(err)
			return
		}

		ABI, err := abi.JSON(strings.NewReader(helloworld.HelloWorldABI))
		if err != nil {
			t.Fatal(err)
			return
		}

		data, err := ABI.Pack("setMessage", "heyho")
		if err != nil {
			t.Fatal(err)
			return
		}

		callQuery := fmt.Sprintf("/tx?to=%s&amount=%d&gasPrice=%d", deployReceipt.ContractAddress.Hex(), 0, 1)
		callReq, err := http.NewRequest("POST", callQuery, bytes.NewBufferString(common.Bytes2Hex(data)))
		if err != nil {
			t.Fatal(err)
			return
		}

		txh := txHandler(client, signer, rules, owner.From, ownerKey)

		callRR := httptest.NewRecorder()
		txh.ServeHTTP(callRR, callReq)

		client.Commit()

		if callRR.Code != 200 {
			t.Errorf("response code = %v, want %v", callRR.Code, 200)
			return
		}

		callReceipt, err := client.TransactionReceipt(ctx, common.HexToHash(callRR.Body.String()))
		if err != nil {
			t.Fatal(err)
			return
		}

		j, _ := callReceipt.MarshalJSON()

		fmt.Println(string(j))

		if len(callReceipt.Logs) != 1 {
			t.Errorf("response code = %v, want %v", len(callReceipt.Logs), 1)
			return
		}
	})
}
