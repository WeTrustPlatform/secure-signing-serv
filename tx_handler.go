package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"sync/atomic"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func txHandler(client Client, signer types.Signer, rules string, owner common.Address, key *ecdsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var to *common.Address
		toString := r.Form.Get("to")
		if toString != "" {
			address := common.HexToAddress(toString)
			to = &address
		}

		valueString := r.Form.Get("value")
		if valueString == "" {
			valueString = "0"
		}
		value := new(big.Int)
		value, ok := value.SetString(valueString, 10)
		if !ok {
			http.Error(w, "Couldn't convert value to big.Int", http.StatusBadRequest)
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
		data = common.Hex2Bytes(string(data))

		gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
			From:  owner,
			To:    to,
			Value: value,
			Data:  data,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		n := atomic.LoadUint64(&nonce)
		tx := &types.Transaction{}
		if toString != "" {
			tx = types.NewTransaction(n, *to, value, gas, gp, data)
		} else {
			tx = types.NewContractCreation(n, big.NewInt(0), gas, gp, data)
		}

		fmt.Println(tx)

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
