package main

import (
	"context"
	"crypto/ecdsa"
	"io/ioutil"
	"math/big"
	"net/http"
	"sync/atomic"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func deployHandler(client Client, signer types.Signer, rules string, owner common.Address, key *ecdsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		gp, ok := big.NewInt(0).SetString(r.Form.Get("gasPrice"), 10)
		if !ok {
			http.Error(w, "Couldn't convert gasPrice to big.Int", http.StatusBadRequest)
			return
		}

		if r.Body == nil {
			http.Error(w, "Request body is mandatory for contract creation", http.StatusBadRequest)
			return
		}
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		data = common.Hex2Bytes(string(data))

		gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
			From: owner,
			To:   nil,
			Data: data,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		n := atomic.LoadUint64(&nonce)
		tx := types.NewContractCreation(n, big.NewInt(0), gas, gp, data)

		valid, err := validate(rules, tx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Error(w, "Invalid transaction", http.StatusUnauthorized)
			return
		}

		signedTx, err := types.SignTx(tx, signer, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		atomic.AddUint64(&nonce, 1)

		w.Write([]byte(signedTx.Hash().String()))
	}
}
