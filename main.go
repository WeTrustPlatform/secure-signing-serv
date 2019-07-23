package main

import (
	"math/big"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func main() {
	for _, v := range []string{"RPC_ENDPOINT", "PRIV_KEY", "PASSPHRASE", "PORT", "BASIC_AUTH_USER", "BASIC_AUTH_PASS", "CHAIN_ID"} {
		if os.Getenv(v) == "" {
			panic("Environment variable not set: " + v)
		}
	}

	client, err := ethclient.Dial(os.Getenv("RPC_ENDPOINT"))
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&transaction{})

	keyJSON := os.Getenv("PRIV_KEY")
	pass := os.Getenv("PASSPHRASE")
	key, err := keystore.DecryptKey([]byte(keyJSON), pass)
	if err != nil {
		panic(err)
	}

	rules := os.Getenv("RULES")

	chainID, ok := big.NewInt(0).SetString(os.Getenv("CHAIN_ID"), 10)
	if !ok {
		panic("Can't parse CHAIN_ID")
	}
	signer := types.NewEIP155Signer(chainID)

	http.HandleFunc("/v1/proxy/transactions", basicAuth(handler(
		client,
		signer,
		rules,
		key.Address,
		key.PrivateKey,
		db)))

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}
