package main

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"net/http"
	"os"
	"strconv"

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
}

func txHandler(ctx context.Context, client ClientFaker, owner common.Address, key *ecdsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		gl, err := strconv.ParseUint(r.Form.Get("gasLimit"), 10, 64)
		if err != nil {
			http.Error(w, "Couldn't parse gasLimit to uint64", 400)
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

		tx := types.NewTransaction(nonce, t, a, gl, gp, d)

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

	ctx := context.Background()

	http.HandleFunc("/tx", basicAuth(txHandler(
		ctx,
		client,
		key.Address,
		key.PrivateKey)))

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
