package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/WeTrustPlatform/secure-signing-serv/client"
	"github.com/ethereum/go-ethereum/common"
)

func main() {
	var endpoint, value, gasPrice, data string
	flag.StringVar(&endpoint, "E", "", "The S3 API endpoint")
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

	s3 := client.NewClient(endpoint)
	resp, err := s3.Transact(
		common.HexToAddress("0xC7f965a58942dbf4E9fbdf77A511863d7041339d"),
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
