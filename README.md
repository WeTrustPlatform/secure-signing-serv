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

Sending a simple transaction:

    curl -X POST "https://your-api-key:your-api-secret@domain-name.com/tx/to-address/amount/gas-limit/gas-price/data-field"

`to-address` is the ethereum address of the receiver (for example `0x5597285BbE81BaF351e2C0884e9a5f4416958862`)
`amount` is the amount transfered to the receiver in wei (for example `1000000000000000`)
`gas-limit` is the maximum number of gas (for example `21000`)
`ges-price` is how much wei you're ready to pay per gas unit (for example `20000000000`)
