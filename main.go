package main

import (
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	log "github.com/sirupsen/logrus"
)

func main() {
	logInit()

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

	http.HandleFunc("/v1/proxy/transactions", basicAuth(txHandler(
		client,
		signer,
		rules,
		key.Address,
		key.PrivateKey,
		db)))

	http.HandleFunc("/v1/proxy/transactions/retry", basicAuth(retryHandler(
		client,
		signer,
		rules,
		key.Address,
		key.PrivateKey,
		db)))

	log.WithFields(log.Fields{
		"PORT": os.Getenv("PORT"),
	}).Info("Listening")

	http.ListenAndServe(":"+os.Getenv("PORT"), nil)
}

func logInit() {
	// Setup log formatter. Default is JSON for production and staging.
	switch os.Getenv("LOG_FORMATTER") {
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	default:
		log.SetFormatter(&log.JSONFormatter{})
	}

	// Only log the warning severity or above.
	logLevel := os.Getenv("LOG_LEVEL")
	logLevelMap := map[string]log.Level{
		"":      log.InfoLevel,
		"panic": log.PanicLevel,
		"fatal": log.FatalLevel,
		"error": log.ErrorLevel,
		"warn":  log.WarnLevel,
		"info":  log.InfoLevel,
		"debug": log.DebugLevel,
		"trace": log.TraceLevel,
	}
	log.SetLevel(logLevelMap[strings.ToLower(logLevel)])
}
