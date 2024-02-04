TODO
====

- [x] bundler-node general setup
- [x] bundler-node libp2p support with peer discovery + prioritization of
peers running the same bundler-node app
- [x] bundler-node gossipsub support to distribute messages across peers
- [x] private key hd wallet support 
- [x] p2p package, node package

- [x] 5189 Operation payload

- [ ] endorser, isOperationReady, check readiness before publishing to mempool. By publishing the
operation to the pool, it means the node checked readiness first. And any pubished message is auto-signed.
Therefore, if a peer continues to send isOperationReady messages which are determined to not be ready,
other peers will begin to score it low and reject its messages.

- [ ] mempool operation queue, as well track the peer/address who sent the operation into the pool for
future peer scoring.

- [ ] chain provider (ethrpc) -- for now, we'll make it work for just a single chain, but really we can support multi-chain

- [ ] ethgas monitoring for pricing

- [ ] sender(s) reading operations from the queue and dispatching native txns to the chain

- [ ] gas estimation for operation -- how to do this for different smart wallets tho...?
- [ ] compression -- how to know which compression module to use, etc. prob endorser / similar can tell us ..
- [ ] endorser should tell us smart wallet version, ya? so we can determine if we can bundle, etc..?
- [ ] how will bundling work for arbitrary smart wallets..?



Other
=====

- [ ] check ethereum/prsym p2p -- does it use gossipsub? mempool..? protobuf..?
- [ ] mempool -- get messages and add to mempool..? what if bundler wants to censor..?
- [ ] how do eth / prysm nodes prevent censorship..? 
- [ ] mempool -- which bundler will actually process the operation..? is it a race?
we dont want multiple bundlers sending out the same information for no reason


Features
========

- [ ] fee token re-payment (standard..)
- [ ] sponsored gas (ie. payment by gas tank, etc)
