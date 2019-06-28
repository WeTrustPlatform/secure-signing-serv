# Secure Signing Service (3S)

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

    curl -X POST "https://your-api-key:your-api-secret@domain-name.com/tx?to=0x5597285BbE81BaF351e2C0884e9a5f4416958862&amount=1000000000000000&gasPrice=20000000000" -H "Content-Type: text/plain" --data hello

Deploying a smart contract:

    curl -X POST "https://your-api-key:your-api-secret@domain-name.com/deploy?gasPrice=20000000000" -H "Content-Type: text/plain" --data 608060405234801561001057600080fd5b50610108806100206000396000f3fe6080604052348015600f57600080fd5b506004361060285760003560e01c8063942ae0a714602d575b600080fd5b603360a5565b6040805160208082528351818301528351919283929083019185019080838360005b83811015606b5781810151838201526020016055565b50505050905090810190601f16801560975780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b60408051808201909152600c81527f68656c6c6f20776f726c6421000000000000000000000000000000000000000060208201529056fea165627a7a7230582092b2d98f6dffec3d6346d5fa88b936f64cfeada8abf83c5339c7f65451dfb6630029
