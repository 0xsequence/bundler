TODO
====

- [x] bundler-node general setup
- [x] bundler-node libp2p support with peer discovery + prioritization of
peers running the same bundler-node app
- [x] bundler-node gossipsub support to distribute messages across peers
- [x] private key hd wallet support 
- [x] p2p package, node package

- [ ] 5189 Operation payload..


Other
=====

- [ ] check ethereum/prsym p2p -- does it use gossipsub? mempool..? protobuf..?
- [ ] mempool -- get messages and add to mempool..? what if bundler wants to censor..?
- [ ] how do eth / prysm nodes prevent censorship..? 
- [ ] mempool -- which bundler will actually process the operation..? is it a race?
we dont want multiple bundlers sending out the same information for no reason
