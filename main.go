package main

import (
	"context"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var nonce uint64

func main() {
	for _, v := range []string{"RPC_ENDPOINT", "PRIV_KEY", "PASSPHRASE", "PORT", "BASIC_AUTH_USER", "BASIC_AUTH_PASS", "CHAIN_ID"} {
		if os.Getenv(v) == "" {
			panic("Environment variable not set: " + v)
		}
	}

	client, err := ethclient.Dial(os.Getenv("RPC_ENDPOINT"))
	if err != nil {
		panic(err)
	}

	keyJSON := os.Getenv("PRIV_KEY")
	pass := os.Getenv("PASSPHRASE")
	key, err := keystore.DecryptKey([]byte(keyJSON), pass)
	if err != nil {
		panic(err)
	}

	nonce, err = client.NonceAt(context.Background(), key.Address, nil)
	if err != nil {
		panic(err)
	}

	rules, err := ioutil.ReadFile("rules.lua")
	if err != nil {
		panic(err)
	}

	chainID, ok := big.NewInt(0).SetString(os.Getenv("CHAIN_ID"), 10)
	if !ok {
		panic("Can't parse CHAIN_ID")
	}
	signer := types.NewEIP155Signer(chainID)

	http.HandleFunc("/v1/proxy/transactions", basicAuth(handler(
		client,
		signer,
		string(rules),
		key.Address,
		key.PrivateKey)))

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
