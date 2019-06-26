# secure-signing-serv

A micro service to store an ethereum key and perform whitelisted transactions on your behalf

## Building

    export BASIC_AUTH_USER=your-api-key
    export BASIC_AUTH_PASS=your-api-secret
    export PASSPHRASE=your-ethereum-account-passphrase
    export PRIV_KEY='{"address":"634f5c3f019f2d44341c3922230bdad2e91e1d9f",...,"version":3}'
    export RPC_ENDPOINT='https://rinkeby.infura.io/v3/your-infura-secret'
    go build
    ./secure-signing-serv

## Usage

    curl -X POST "https://your-api-key:your-api-secret@domain-name.com/tx/to-address/amount-in-wei/gas-limit/gas-price/data-field"
