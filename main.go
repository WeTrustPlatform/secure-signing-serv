package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client allow passing an ethclient.Client or a backend.SimulatedBackend
type Client interface {
	ethereum.ChainStateReader
	ethereum.TransactionSender
	ethereum.GasEstimator
}

var calls chan ethereum.CallMsg

func process(call ethereum.CallMsg, client *ethclient.Client, rules string, signer types.Signer, key *keystore.Key) {
	ctx := context.Background()

	nonce, err := client.NonceAt(ctx, call.From, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	gas, err := client.EstimateGas(ctx, call)
	if err != nil {
		fmt.Println(err)
		return
	}

	tx := types.NewTransaction(nonce, *call.To, call.Value, gas, call.GasPrice, call.Data)

	valid, err := validate(rules, tx)
	if err != nil {
		fmt.Println(err)
		return
	}
	if !valid {
		fmt.Println("Invalid transaction")
		return
	}

	signedTx, err := types.SignTx(tx, signer, key.PrivateKey)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		time.Sleep(1 * time.Second)
		receipt, _ := client.TransactionReceipt(ctx, signedTx.Hash())
		if receipt != nil {
			break
		}
	}

	fmt.Println(signedTx.Hash().String())
}

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

	rules, err := ioutil.ReadFile("rules.lua")
	if err != nil {
		panic(err)
	}

	chainID, ok := big.NewInt(0).SetString(os.Getenv("CHAIN_ID"), 10)
	if !ok {
		panic("Can't parse CHAIN_ID")
	}
	signer := types.NewEIP155Signer(chainID)

	calls = make(chan ethereum.CallMsg, 16)

	http.HandleFunc("/tx", basicAuth(txHandler(
		client,
		signer,
		string(rules),
		key.Address,
		key.PrivateKey)))

	http.HandleFunc("/deploy", basicAuth(deployHandler(
		client,
		signer,
		string(rules),
		key.Address,
		key.PrivateKey)))

	go func() {
		for {
			process(<-calls, client, string(rules), signer, key)
		}
	}()

	fmt.Println("Listening...")
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
