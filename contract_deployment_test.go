package main

import (
	"bytes"
	"context"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/WeTrustPlatform/secure-signing-serv/sss"
	"github.com/WeTrustPlatform/secure-signing-serv/testdata/helloworld"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

func Test_contractDeployment(t *testing.T) {
	ctx := context.Background()

	ownerKey, _ := crypto.GenerateKey()
	owner := bind.NewKeyedTransactor(ownerKey)

	rules := `function validate(tx) return true end`
	signer := types.HomesteadSigner{}

	t.Run("Can deploy a contract", func(t *testing.T) {
		client := backends.NewSimulatedBackend(core.GenesisAlloc{
			owner.From: core.GenesisAccount{Balance: big.NewInt(50000000000)},
		}, 4000000)

		byteCode := helloworld.HelloWorldBin[2:]

		p := sss.TxPayload{GasPrice: "1", Data: byteCode}
		b := new(bytes.Buffer)
		json.NewEncoder(b).Encode(p)
		req, err := http.NewRequest("POST", "/tx", b)
		if err != nil {
			t.Fatal(err)
			return
		}

		h := txHandler(client, signer, rules, owner.From, ownerKey, &dbMock{})
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
		if receipt == nil {
			t.Errorf("transaction receipt = %v", nil)
			return
		}

		codeAtAddress, err := client.CodeAt(ctx, receipt.ContractAddress, nil)
		if err != nil {
			t.Fatal(err)
			return
		}

		// Why the output is different than byteCode? We'll have to figure out
		want := `608060405234801561001057600080fd5b50600436106100365760003560e01c8063368b87721461003b578063e21f37ce146100e3575b600080fd5b6100e16004803603602081101561005157600080fd5b81019060208101813564010000000081111561006c57600080fd5b82018360208201111561007e57600080fd5b803590602001918460018302840111640100000000831117156100a057600080fd5b91908080601f016020809104026020016040519081016040528093929190818152602001838380828437600092019190915250929550610160945050505050565b005b6100eb610210565b6040805160208082528351818301528351919283929083019185019080838360005b8381101561012557818101518382015260200161010d565b50505050905090810190601f1680156101525780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b805161017390600090602084019061029e565b50604080516020808252835181830152835133937f2ef3fc8a662077a0e040f48a65ffa7573c31f49d3f910d11faaebafb4024c6529386939092839283019185019080838360005b838110156101d35781810151838201526020016101bb565b50505050905090810190601f1680156102005780820380516001836020036101000a031916815260200191505b509250505060405180910390a250565b6000805460408051602060026001851615610100026000190190941693909304601f810184900484028201840190925281815292918301828280156102965780601f1061026b57610100808354040283529160200191610296565b820191906000526020600020905b81548152906001019060200180831161027957829003601f168201915b505050505081565b828054600181600116156101000203166002900490600052602060002090601f016020900481019282601f106102df57805160ff191683800117855561030c565b8280016001018555821561030c579182015b8281111561030c5782518255916020019190600101906102f1565b5061031892915061031c565b5090565b61033691905b808211156103185760008155600101610322565b9056fea165627a7a72305820a9ef249bf31d8f35838cabe1e58664dd84b47baf7830cda567c5e7ae48ee5b320029`
		if common.Bytes2Hex(codeAtAddress) != want {
			t.Errorf("code at address = %v, want %v", common.Bytes2Hex(codeAtAddress), want)
			return
		}
	})
}
