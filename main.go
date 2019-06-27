package main

import (
	"context"
	"crypto/ecdsa"
	"io/ioutil"
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ClientFaker allow passing an ethclient.Client or a backend.SimulatedBackend
type ClientFaker interface {
	ethereum.ChainStateReader
	ethereum.TransactionSender
	ethereum.GasEstimator
}

func txHandler(client ClientFaker, owner common.Address, key *ecdsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), 400)
		}

		t := common.HexToAddress(r.Form.Get("to"))

		a := new(big.Int)
		a, ok := a.SetString(r.Form.Get("amount"), 10)
		if !ok {
			http.Error(w, "Couldn't convert amount to big.Int", 400)
			return
		}

		gp, ok := big.NewInt(0).SetString(r.Form.Get("gasPrice"), 10)
		if !ok {
			http.Error(w, "Couldn't convert gasPrice to big.Int", 400)
			return
		}

		d := []byte(r.Form.Get("data"))

		nonce, err := client.NonceAt(ctx, owner, nil)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
			From:  owner,
			To:    &t,
			Value: a,
			Data:  d,
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		tx := types.NewTransaction(nonce, t, a, gas, gp, d)

		rules, err := ioutil.ReadFile("rules.lua")
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		valid, err := validate(string(rules), tx)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		if !valid {
			http.Error(w, "Invalid transaction", 401)
			return
		}

		signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, key)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		w.Write([]byte(signedTx.Hash().String()))
	}
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

	http.HandleFunc("/tx", basicAuth(txHandler(
		client,
		key.Address,
		key.PrivateKey)))

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
