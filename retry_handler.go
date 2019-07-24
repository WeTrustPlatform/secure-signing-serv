package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"net/http"
	"strings"

	"github.com/WeTrustPlatform/secure-signing-serv/sss"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	log "github.com/sirupsen/logrus"
)

func retryHandler(
	client Client,
	signer types.Signer,
	rules string,
	owner common.Address,
	key *ecdsa.PrivateKey,
	db Recorder,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		if r.Method != "PATCH" {
			http.Error(w, "retry only supports PATCH method", http.StatusMethodNotAllowed)
			return
		}

		hash := strings.TrimPrefix(r.URL.Path, "/v1/proxy/transactions/")

		decoder := json.NewDecoder(r.Body)
		var p sss.RetryPayload
		err := decoder.Decode(&p)
		if err != nil {
			http.Error(w, "error decoding payload: "+err.Error(), http.StatusBadRequest)
			return
		}

		if len(p) != 1 || p[0].Op != "replace" || p[0].Path != "/gasPrice" {
			http.Error(w, "only replacing gasPrice is supported", http.StatusBadRequest)
			return
		}

		gp, ok := big.NewInt(0).SetString(p[0].Value, 10)
		if !ok {
			http.Error(w, "couldn't convert gasPrice to big.Int", http.StatusBadRequest)
			return
		}

		oldTx := transaction{}
		db.First(&oldTx, "hash = ?", hash)

		var tx *types.Transaction
		if oldTx.To != "" {
			value, _ := big.NewInt(0).SetString(oldTx.Value, 10)
			tx = types.NewTransaction(
				oldTx.Nonce,
				common.HexToAddress(oldTx.To),
				value,
				oldTx.Gas,
				gp,
				common.Hex2Bytes(oldTx.Data),
			)
		} else {
			tx = types.NewContractCreation(
				oldTx.Nonce,
				big.NewInt(0),
				oldTx.Gas,
				gp,
				common.Hex2Bytes(oldTx.Data),
			)
		}

		valid, err := validate(rules, tx)
		if err != nil {
			http.Error(w, "error validating transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if !valid {
			log.WithFields(log.Fields{
				"Nonce":    oldTx.Nonce,
				"To":       oldTx.To,
				"Value":    oldTx.Value,
				"Gas":      oldTx.Gas,
				"GasPrice": gp.String(),
				"Hash":     tx.Hash().String(),
			}).Warning("Forbidden transaction")
			http.Error(w, "forbidden transaction", http.StatusForbidden)
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

		oldTx.Hash = signedTx.Hash().String()
		oldTx.GasPrice = gp.String()
		db.Save(oldTx)

		log.WithFields(log.Fields{
			"Nonce":    oldTx.Nonce,
			"To":       oldTx.To,
			"Value":    oldTx.Value,
			"Gas":      oldTx.Gas,
			"GasPrice": gp.String(),
			"Hash":     signedTx.Hash().String(),
		}).Info("Successfully retried transaction")

		w.Write([]byte(signedTx.Hash().String()))
	}
}
