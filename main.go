package main

import (
	"io/ioutil"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client allow passing an ethclient.Client or a backend.SimulatedBackend
type Client interface {
	ethereum.ChainStateReader
	ethereum.TransactionSender
	ethereum.GasEstimator
}

func main() {
	for _, v := range []string{"RPC_ENDPOINT", "PRIV_KEY", "PASSPHRASE", "PORT", "BASIC_AUTH_USER", "BASIC_AUTH_PASS"} {
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

	rules, err := ioutil.ReadFile("rules.lua")
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/tx", basicAuth(txHandler(
		client,
		rules,
		key.Address,
		key.PrivateKey)))

	http.HandleFunc("/deploy", basicAuth(deployHandler(
		client,
		rules,
		key.Address,
		key.PrivateKey)))

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
