package client

import (
	"bytes"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
)

// Client is a helper to perform transactions and contract deployments using S3
// over HTTP
type Client struct {
	Endpoint string
}

// NewClient instantiates an S3 client to query the given endpoint
func NewClient(e string) *Client {
	return &Client{
		Endpoint: e,
	}
}

// Transact performs a transaction or a contract call
func (c *Client) Transact(to common.Address, value, gasPrice *big.Int, data string) (*http.Response, error) {
	query := fmt.Sprintf("%s/tx?to=%s&value=%s&gasPrice=%s", c.Endpoint, to.Hex(), value.String(), gasPrice.String())
	return http.Post(query, "text/plain", bytes.NewBufferString(data))
}

// Deploy deploys a smart contract
func (c *Client) Deploy(value, gasPrice *big.Int, data string) (*http.Response, error) {
	query := fmt.Sprintf("%s/deploy?value=%s&gasPrice=%s", c.Endpoint, value.String(), gasPrice.String())
	return http.Post(query, "text/plain", bytes.NewBufferString(data))
}
