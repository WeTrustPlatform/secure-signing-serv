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

// Transact performs a transaction or a contract call
func (c *Client) Transact(to common.Address, value, gasPrice *big.Int, data string) (*http.Response, error) {
	p := Payload{
		To:       to.Hex(),
		Value:    value.String(),
		GasPrice: gasPrice.String(),
		Data:     data,
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(p)
	return http.Post(c.Endpoint+"/tx", "application/json", b)
}
