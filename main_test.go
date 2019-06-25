package main

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

func Test_txHandler(t *testing.T) {
	ctx := context.Background()

	ownerKey, _ := crypto.GenerateKey()
	owner := bind.NewKeyedTransactor(ownerKey)
	ownerAuth := bind.NewKeyedTransactor(ownerKey)

	client := backends.NewSimulatedBackend(core.GenesisAlloc{
		owner.From: core.GenesisAccount{Balance: big.NewInt(50000000000)},
	}, 4712388)

	h := txHandler(ctx, client, ownerAuth, ownerKey)

	client.Commit()

	t.Run("Can proxy a simple tx", func(t *testing.T) {
		// if got := txHandler(tt.args.ctx, tt.args.client, tt.args.key); !reflect.DeepEqual(got, tt.want) {
		// 	t.Errorf("txHandler() = %v, want %v", got, tt.want)
		// }
		fmt.Println(h)
	})
}
