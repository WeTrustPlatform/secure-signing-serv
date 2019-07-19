package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/WeTrustPlatform/secure-signing-serv/sss"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	var endpoint, to, value, gasPrice, data string
	flag.StringVar(&endpoint, "endpoint", "", "The 3S API endpoint")
	flag.StringVar(&to, "to", "", "The receiver address")
	flag.StringVar(&value, "value", "0", "The amount to be transfered")
	flag.StringVar(&gasPrice, "gasprice", "0", "The price of the gas")
	flag.StringVar(&data, "data", "", "Data field of the transaction")
	flag.Parse()

	v, ok := big.NewInt(0).SetString(value, 10)
	if !ok {
		fmt.Println("Couldn't parse value")
		return
	}
	gp, ok := big.NewInt(0).SetString(gasPrice, 10)
	if !ok {
		fmt.Println("Couldn't parse gasPrice")
		return
	}

	var toAddrPtr *common.Address
	if to != "" {
		addr := common.HexToAddress(to)
		toAddrPtr = &addr
	}

	c := sss.NewClient(endpoint)
	resp, err := c.Transact(
		toAddrPtr,
		v,
		gp,
		data,
	)
	if err != nil {
		fmt.Println(err)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(resp.StatusCode, string(body))
}
