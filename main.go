package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client allow passing an ethclient.Client or a backend.SimulatedBackend
type Client interface {
	ethereum.ChainStateReader
	ethereum.TransactionSender
	ethereum.TransactionReader
	ethereum.GasEstimator
}

var calls chan ethereum.CallMsg

func init() {
	calls = make(chan ethereum.CallMsg, 2048)
}

func process(call ethereum.CallMsg, client Client, rules string, signer types.Signer, key *ecdsa.PrivateKey) error {
	fmt.Println(call)
	ctx := context.Background()

	nonce, err := client.NonceAt(ctx, call.From, nil)
	if err != nil {
		return err
	}

	gas, err := client.EstimateGas(ctx, call)
	if err != nil {
		return err
	}

	tx := types.NewTransaction(nonce, *call.To, call.Value, gas, call.GasPrice, call.Data)

	signedTx, err := types.SignTx(tx, signer, key)
	if err != nil {
		return err
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return err
	}

	// for {
	// 	time.Sleep(1 * time.Second)
	// 	receipt, _ := client.TransactionReceipt(ctx, signedTx.Hash())
	// 	if receipt != nil {
	// 		break
	// 	}
	// }

	fmt.Println(signedTx.Hash().String())
	return nil
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
			err := process(<-calls, client, string(rules), signer, key.PrivateKey)
			if err != nil {
				fmt.Print(err)
			}
		}
	}()

	fmt.Println("Listening...")
	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
