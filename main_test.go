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

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)

		client.Commit()

		if rr.Code != 200 {
			t.Errorf("response code = %v, want %v", rr.Code, 200)
		}

		deployReceipt, err := client.TransactionReceipt(ctx, common.HexToHash(rr.Body.String()))
		if err != nil {
			t.Fatal(err)
		}

		testabi, err := abi.JSON(strings.NewReader(helloworld.HelloWorldABI))
		if err != nil {
			t.Fatal(err)
		}

		bytesData, err := testabi.Pack("setMessage", "This is a test")
		if err != nil {
			t.Fatal(err)
		}

		callQuery := fmt.Sprintf("/tx?to=%s&amount=%d&gasPrice=%d", deployReceipt.ContractAddress.Hex(), 0, 1)
		txReq, err := http.NewRequest("POST", callQuery, bytes.NewBufferString(common.Bytes2Hex(bytesData)))
		if err != nil {
			t.Fatal(err)
		}

		txh := txHandler(client, signer, rules, owner.From, ownerKey)

		callRR := httptest.NewRecorder()
		txh.ServeHTTP(rr, txReq)

		client.Commit()

		if callRR.Code != 200 {
			t.Errorf("response code = %v, want %v", callRR.Code, 200)
		}

		callReceipt, err := client.TransactionReceipt(ctx, common.HexToHash(rr.Body.String()))
		if err != nil {
			t.Fatal(err)
		}

		b, _ := callReceipt.MarshalJSON()

		fmt.Println(callReceipt)
		fmt.Println(string(b))
	})
}
