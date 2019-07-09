package main

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WeTrustPlatform/secure-signing-serv/sss"
	"github.com/WeTrustPlatform/secure-signing-serv/testdata/helloworld"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func Test_contractCall(t *testing.T) {
	ctx := context.Background()

	ownerKey, _ := crypto.GenerateKey()
	owner := bind.NewKeyedTransactor(ownerKey)

	rules := `function validate(tx) return true end`
	signer := types.HomesteadSigner{}

	t.Run("Can call a method of a smart contract", func(t *testing.T) {
		client := backends.NewSimulatedBackend(core.GenesisAlloc{
			owner.From: core.GenesisAccount{Balance: big.NewInt(50000000000)},
		}, 4000000)
		message := "hohay"

		byteCode := helloworld.HelloWorldBin[2:]

		p := sss.Payload{GasPrice: "1", Data: byteCode}
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(p)
		req, err := http.NewRequest("POST", "/tx", b)
		if err != nil {
			t.Fatal(err)
			return
		}

		h := handler(client, signer, rules, owner.From, ownerKey)
		deployRR := httptest.NewRecorder()
		h.ServeHTTP(deployRR, req)
		client.Commit()

		if deployRR.Code != 200 {
			t.Errorf("response code = %v, want %v", deployRR.Code, 200)
			return
		}

		deployReceipt, err := client.TransactionReceipt(ctx, common.HexToHash(deployRR.Body.String()))
		if err != nil {
			t.Fatal(err)
			return
		}
		if deployReceipt == nil {
			t.Errorf("transaction receipt = %v", nil)
			return
		}

		ABI, err := abi.JSON(strings.NewReader(helloworld.HelloWorldABI))
		if err != nil {
			t.Fatal(err)
			return
		}

		data, err := ABI.Pack("setMessage", message)
		if err != nil {
			t.Fatal(err)
			return
		}

		p2 := sss.Payload{To: deployReceipt.ContractAddress.Hex(), GasPrice: "1", Data: common.Bytes2Hex(data)}
		b2 := new(bytes.Buffer)
		json.NewEncoder(b2).Encode(p2)
		callReq, err := http.NewRequest("POST", "/tx", b2)
		if err != nil {
			t.Fatal(err)
			return
		}

		txh := handler(client, signer, rules, owner.From, ownerKey)
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
		if callReceipt == nil {
			t.Errorf("transaction receipt = %v", nil)
			return
		}

		if len(callReceipt.Logs) != 1 {
			t.Errorf("response code = %v, want %v", len(callReceipt.Logs), 1)
			return
		}

		contract, err := helloworld.NewHelloWorld(deployReceipt.ContractAddress, client)
		if err != nil {
			t.Fatal(err)
			return
		}

		m, err := contract.Message(&bind.CallOpts{})
		if err != nil {
			t.Fatal(err)
			return
		}

		if m != message {
			t.Errorf("message = %v, want %v", m, message)
			return
		}
	})
}
