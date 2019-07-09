package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"net/http"
	"sync/atomic"

	"github.com/WeTrustPlatform/secure-signing-serv/sss"
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

func handler(client Client, signer types.Signer, rules string, owner common.Address, key *ecdsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		ctx := context.Background()

		decoder := json.NewDecoder(r.Body)
		var p sss.Payload
		err := decoder.Decode(&p)
		if err != nil {
			http.Error(w, "error decoding payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		var to *common.Address
		if p.To != "" {
			address := common.HexToAddress(p.To)
			to = &address
		}

		value := new(big.Int)
		if p.Value != "" {
			value, ok = value.SetString(p.Value, 10)
			if !ok {
				http.Error(w, "couldn't convert value to big.Int", http.StatusBadRequest)
				return
			}
		}

		gp, ok := big.NewInt(0).SetString(p.GasPrice, 10)
		if !ok {
			http.Error(w, "couldn't convert gasPrice to big.Int", http.StatusBadRequest)
			return
		}

		data := common.Hex2Bytes(p.Data)

		gas, err := client.EstimateGas(ctx, ethereum.CallMsg{
			From:  owner,
			To:    to,
			Value: value,
			Data:  data,
		})
		if err != nil {
			http.Error(w, "error estimating gas: "+err.Error(), http.StatusInternalServerError)
			return
		}

		n := atomic.LoadUint64(&nonce)
		var tx *types.Transaction
		if to != nil {
			tx = types.NewTransaction(n, *to, value, gas, gp, data)
		} else {
			tx = types.NewContractCreation(n, big.NewInt(0), gas, gp, data)
		}

		valid, err := validate(rules, tx)
		if err != nil {
			http.Error(w, "error validating transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if !valid {
			http.Error(w, "transaction forbidden", http.StatusForbidden)
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
