package sss

import (
	"bytes"
	"encoding/json"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
)

// Client is a helper to perform transactions and contract deployments using S3
// over HTTP
type Client struct {
	Endpoint string
}

// NewClient instantiates an 3S client to query the given endpoint
func NewClient(e string) *Client {
	return &Client{
		Endpoint: e,
	}
}

// Transact performs a transaction, a contract deployment or a contract call
func (c *Client) Transact(to *common.Address, value, gasPrice *big.Int, data string) (*http.Response, error) {
	toStr := ""
	if to != nil {
		toStr = to.Hex()
	}
	p := TxPayload{
		To:       toStr,
		Value:    value.String(),
		GasPrice: gasPrice.String(),
		Data:     data,
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(p)
	return http.Post(c.Endpoint+"/v1/proxy/transactions", "application/json", b)
}

// Retry retries a failed transaction with a different gas price
func (c *Client) Retry(hash common.Hash, gasPrice *big.Int) (*http.Response, error) {
	p := RetryPayload{
		Hash:     hash.Hex(),
		GasPrice: gasPrice.String(),
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(p)
	return http.Post(c.Endpoint+"/v1/proxy/transactions/retry", "application/json", b)
}
