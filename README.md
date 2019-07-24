# Secure Signing Service (3S)

A micro service to store an ethereum key and perform whitelisted transactions on your behalf. S3 can transfer ether, deploy smart contracts, and call smart contract methods.

It is useful when you want your users to be able to write in the ethereum blockchain without having to pay the gas fees. You pay the fee for them.

The service is authenticated, and scriptable using Lua to only allow certain kind of transactions.

## Building

    export BASIC_AUTH_USER=your-api-key
    export BASIC_AUTH_PASS=your-api-secret
    export PASSPHRASE=your-ethereum-account-passphrase
    export PRIV_KEY='{"address":"634f5c3f019f2d44341c3922230bdad2e91e1d9f",...,"version":3}'
    export RPC_ENDPOINT='https://rinkeby.infura.io/v3/your-infura-secret'
    export CHAIN_ID=4
    export RULES='function validate(tx) return true end'
    go build
    ./secure-signing-serv

## Usage

Sending a simple transaction:

    curl -X POST "https://your-api-key:your-api-secret@domain-name.com/v1/proxy/transactions" -H "Content-Type: application/json" --data '{"to":"0x5597285BbE81BaF351e2C0884e9a5f4416958862","value":"1000000000000000","gasPrice":"20000000000"}'

Deploying a smart contract:

    curl -X POST "https://your-api-key:your-api-secret@domain-name.com/v1/proxy/transactions" -H "Content-Type: application/json" --data '{"gasPrice":"20000000000","data":"6080...0029"}'

Calling a smart contract method:

    curl -X POST "http://your-api-key:your-api-secret@localhost:3005/v1/proxy/transactions" -H "Content-Type: application/json" --data '{"to":"0xC7f965a58942dbf4E9fbdf77A511863d7041339d","gasPrice":"40000000000","data":"368b877200000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000008626172626f757365000000000000000000000000000000000000000000000000"}'

Retrying a transaction with higher gas price:

    curl -X PATCH "http://your-api-key:your-api-secret@localhost:3005/v1/proxy/transactions/0x9a08c7c158b6d0bdf43cd2af6e78892edf01e25a4306f3063ab610a48d0e5c0b" -H "Content-Type: application/json" --data '{"op":"replace","path":"/gasPrice","value":"20000000000"}'

## Whitelisting transactions

You can write a Lua script that check the transaction properties to filter transactions.

For example:

    RULES='function validate(tx) return tx.to == "0x5597285BbE81BaF351e2C0884e9a5f4416958862" or tx.value == "10000000000" end'
