package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/WeTrustPlatform/secure-signing-serv/sss"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/jinzhu/gorm"
	log "github.com/sirupsen/logrus"
)

// Client allows passing an ethclient.Client or a backend.SimulatedBackend
type Client interface {
	ethereum.ChainStateReader
	ethereum.TransactionSender
	ethereum.GasEstimator
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
}

// Recorder allows mocking the database operations
type Recorder interface {
	Create(interface{}) *gorm.DB
	First(interface{}, ...interface{}) *gorm.DB
	Save(interface{}) *gorm.DB
}

func txHandler(
	client Client,
	signer types.Signer,
	rules string,
	owner common.Address,
	key *ecdsa.PrivateKey,
	db Recorder,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ok bool
		ctx := context.Background()

		decoder := json.NewDecoder(r.Body)
		var p sss.TxPayload
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
			log.WithFields(log.Fields{
				"To":       p.To,
				"Value":    value.String(),
				"Gas":      gas,
				"GasPrice": gp.String(),
				"error":    err.Error(),
			}).Error("Error estimating gas.")
			// Should return here because the transaction will fail eventually
			http.Error(w, "error estimating gas: "+err.Error(), http.StatusInternalServerError)
			return
		}

		nonce, err := client.PendingNonceAt(ctx, owner)
		if err != nil {
			log.WithFields(log.Fields{
				"Nonce":    nonce,
				"To":       p.To,
				"Value":    value.String(),
				"Gas":      gas,
				"GasPrice": gp.String(),
				"error":    err.Error(),
			}).Error("Error getting pending nonce.")
			http.Error(w, "error getting pending nonce: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var tx *types.Transaction
		if to != nil {
			tx = types.NewTransaction(nonce, *to, value, gas, gp, data)
		} else {
			tx = types.NewContractCreation(nonce, big.NewInt(0), gas, gp, data)
		}

		valid, err := validate(rules, tx)
		if err != nil {
			log.WithFields(log.Fields{
				"Nonce":    nonce,
				"To":       p.To,
				"Value":    value.String(),
				"Gas":      gas,
				"GasPrice": gp.String(),
				"error":    err.Error(),
			}).Error("Error validating transaction.")
			http.Error(w, "error validating transaction: "+err.Error(), http.StatusInternalServerError)
			return
		}
		if !valid {
			log.WithFields(log.Fields{
				"Nonce":    nonce,
				"To":       p.To,
				"Value":    value.String(),
				"Gas":      gas,
				"GasPrice": gp.String(),
				"Hash":     tx.Hash().String(),
			}).Warning("Forbidden transaction")
			http.Error(w, "forbidden transaction", http.StatusForbidden)
			return
		}

		signedTx, err := types.SignTx(tx, signer, key)
		if err != nil {
			// DO NOT log signedTx
			log.WithFields(log.Fields{
				"Nonce":    nonce,
				"To":       p.To,
				"Value":    value.String(),
				"Gas":      gas,
				"GasPrice": gp.String(),
				"Hash":     signedTx.Hash().String(),
				"error":    err.Error(),
			}).Error("Error types.SignTx.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			// DO NOT log signedTx
			log.WithFields(log.Fields{
				"Nonce":    nonce,
				"To":       p.To,
				"Value":    value.String(),
				"Gas":      gas,
				"GasPrice": gp.String(),
				"Hash":     signedTx.Hash().String(),
				"error":    err.Error(),
			}).Error("Error client.sendTransaction.")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		db.Create(&transaction{
			Nonce:    nonce,
			To:       p.To,
			Value:    value.String(),
			Gas:      gas,
			GasPrice: gp.String(),
			Data:     p.Data,
			Hash:     signedTx.Hash().String(),
		})

		log.WithFields(log.Fields{
			"Nonce":    nonce,
			"To":       p.To,
			"Value":    value.String(),
			"Gas":      gas,
			"GasPrice": gp.String(),
			"Hash":     signedTx.Hash().String(),
		}).Info("Successfully forwared transaction")

		w.Write([]byte(signedTx.Hash().String()))
	}
}
