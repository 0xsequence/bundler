TODO
====

- [x] bundler-node general setup
- [x] bundler-node libp2p support with peer discovery + prioritization of
peers running the same bundler-node app
- [x] bundler-node gossipsub support to distribute messages across peers
- [x] private key hd wallet support 
- [x] p2p package, node package

- [ ] 5189 Operation payload.. 

- [ ] endorser, isOperationReady, check readiness before adding to mempool -- should we sign too?

- [ ] mempool operation queue, with signing the messages sent into the mempool..
perhaps there is a way to determine the address from the peer / signature..? maybe..

- [ ] chain provider (ethrpc) -- for now, we'll make it work for just a single chain, but really we can support multi-chain

- [ ] sender(s) reading operations from the queue..

- [ ] ethgas monitoring for pricing

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
