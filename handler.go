package main

import (
	"context"
	"crypto/ecdsa"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Client allow passing an ethclient.Client or a backend.SimulatedBackend
type Client interface {
	ethereum.ChainStateReader
	ethereum.TransactionSender
	ethereum.GasEstimator
}

func txHandler(client Client, owner common.Address, key *ecdsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		to := common.HexToAddress(r.Form.Get("to"))

		amount := new(big.Int)
		amount, ok := amount.SetString(r.Form.Get("amount"), 10)
		if !ok {
			http.Error(w, "Couldn't convert amount to big.Int", http.StatusBadRequest)
			return
		}

		gp, ok := big.NewInt(0).SetString(r.Form.Get("gasPrice"), 10)
		if !ok {
			http.Error(w, "Couldn't convert gasPrice to big.Int", http.StatusBadRequest)
			return
		}

		var data []byte
		if r.Body != nil {
			data, err = ioutil.ReadAll(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			defer r.Body.Close()
		}

		nonce, err := client.NonceAt(ctx, owner, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
			From:  owner,
			To:    &to,
			Value: amount,
			Data:  data,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tx := types.NewTransaction(nonce, to, amount, gas, gp, data)

		rules, err := ioutil.ReadFile("rules.lua")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		valid, err := validate(string(rules), tx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Error(w, "Invalid transaction", http.StatusUnauthorized)
			return
		}

		signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write([]byte(signedTx.Hash().String()))
	}
}
