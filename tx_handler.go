package main

import (
	"crypto/ecdsa"
	"io/ioutil"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func txHandler(client Client, signer types.Signer, rules string, owner common.Address, key *ecdsa.PrivateKey) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		to := common.HexToAddress(r.Form.Get("to"))

		value := new(big.Int)
		value, ok := value.SetString(r.Form.Get("value"), 10)
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

		calls <- ethereum.CallMsg{
			From:     owner,
			To:       &to,
			Value:    value,
			GasPrice: gp,
			Data:     data,
		}
	}
}
