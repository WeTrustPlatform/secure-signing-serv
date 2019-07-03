# Secure Signing Service (3S)

A micro service to store an ethereum key and perform whitelisted transactions on your behalf

## Building

    export BASIC_AUTH_USER=your-api-key
    export BASIC_AUTH_PASS=your-api-secret
    export PASSPHRASE=your-ethereum-account-passphrase
    export PRIV_KEY='{"address":"634f5c3f019f2d44341c3922230bdad2e91e1d9f",...,"version":3}'
    export RPC_ENDPOINT='https://rinkeby.infura.io/v3/your-infura-secret'
    export CHAIN_ID=4
    go build
    ./secure-signing-serv

## Usage

Sending a simple transaction:

    curl -X POST "https://your-api-key:your-api-secret@domain-name.com/tx?to=0x5597285BbE81BaF351e2C0884e9a5f4416958862&value=1000000000000000&gasPrice=20000000000" -H "Content-Type: text/plain" --data hello

Deploying a smart contract:

    curl -X POST "https://your-api-key:your-api-secret@domain-name.com/deploy?gasPrice=20000000000" -H "Content-Type: text/plain" --data 6080...0029

Calling a smart contract method:

    curl -X POST "http://your-api-key:your-api-secret@localhost:3005/tx?to=0xC7f965a58942dbf4E9fbdf77A511863d7041339d&value=0&gasPrice=40000000000" -H "Content-Type: text/plain" --data 368b877200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000008626172626f757365000000000000000000000000000000000000000000000000
