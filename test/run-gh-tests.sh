#!/bin/bash

npm install -g ganache-cli
ganache-cli -h "0.0.0.0" -p 8545 -m "concert load couple harbor equip island argue ramp clarify fence smart topic" -l 999999999999999 &

wget https://binaries.soliditylang.org/linux-amd64/solc-linux-amd64-v0.8.4+commit.c7e474f2
mv solc-linux-amd64-v0.8.4+commit.c7e474f2 solc && chmod +x solc && mv solc /usr/local/bin/solc
solc --version

git clone --depth 1 --branch v0.4.0-rc2 https://github.com/cicizeo/hilo.git
cd hilo
make install
cd ..

CHAIN_ID=hilo-local STAKE_DENOM=uhilo DENOM=uhilo CLEANUP=1 ./test/cosmos/multinode.sh hilod
LORAN_TEST_EVM_RPC="http://0.0.0.0:8545" go test ./test/loran/...
