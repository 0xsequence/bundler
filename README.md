Ethereum P2P Transaction Bundler for Smart Wallets
==================================================

A p2p transaction bundler based on ERC5189, supporting Sequence, Gnosis
and ERC4337 UserOps.

## Usage

1. In one terminal, run: `make run`
  * This will start the node on p2p port 5000 and rpc port 3000
2. In another terminal, run: `make run2`
  * This will start the node on random p2p port and rpc port 4000
3. Send a message to either node rpc endpoint which will then broadcast
to the p2p pubsub channel:

```
curl -H "Content-Type: application/json" http://localhost:3000/rpc/Debug/Broadcast -d '{"message":{"txn":"anything goes here"}}'
```