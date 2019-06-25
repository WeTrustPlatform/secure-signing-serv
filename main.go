package main

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"goji.io"
	"goji.io/pat"
)

// ClientFaker allow passing an ethclient.Client or a backend.SimulatedBackend
type ClientFaker interface {
	ethereum.ChainStateReader
	ethereum.TransactionSender
}

func txHandler(ctx context.Context, client ClientFaker, auth *bind.TransactOpts, key *ecdsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t := common.HexToAddress(pat.Param(r, "to"))

		a, errb := big.NewInt(0).SetString(pat.Param(r, "amount"), 10)
		if errb {
			http.Error(w, "Couldn't convert amount to big.Int", 400)
			return
		}

		gl, err := strconv.ParseUint(pat.Param(r, "gasLimit"), 10, 64)
		if err != nil {
			http.Error(w, "Couldn't parse gasLimit to uint64", 400)
			return
		}

		gp, ok := big.NewInt(0).SetString(pat.Param(r, "gasPrice"), 10)
		if !ok {
			http.Error(w, "Couldn't convert gasPrice to big.Int", 400)
			return
		}

		d := []byte(pat.Param(r, "data"))

		nonce, err := client.NonceAt(ctx, auth.From, nil)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		tx := types.NewTransaction(nonce, t, a, gl, gp, d)

		signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, key)

		err = client.SendTransaction(ctx, signedTx)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

func main() {
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

	auth, err := bind.NewTransactor(strings.NewReader(string(keyJSON[:])), pass)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	mux := goji.NewMux()
	mux.Use(basicAuth)

	mux.HandleFunc(pat.Post("/tx/:to/:amount/:gasLimit/:gasPrice/:data"), txHandler(
		ctx,
		client,
		auth,
		key.PrivateKey))

	http.ListenAndServe(":"+os.Getenv("PORT"), mux)
}
