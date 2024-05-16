# Entrypoint and mock environment

## Deployment

Create `.env` with the following:

```
PRIVATE_KEY=0x...
```

Run the deploy command via make:

```
make deploy
```

Contracts will be deployed to:

| Contract             | Address                                      |
| -------------------- | -------------------------------------------- |
| `OperationValidator` | `0x6CDDB903C49CF49e52b853C7a83F6A79E17a0BaA` |
| `TemporalRegistry`   | `0xcd4e127b83e6a170195e6a561eab02406ec8b941` |

## Find Uniswap Pool

The bundler's collector can be configured to require certain fees to be paid in a specific token. This is done by setting the `collector.references` field in the config file. The bundler will then only accept operations that pay the required fee in the required token.

The collector uses the Uniswap v2 pool to find the current price of the token to be accepted. This pool should be a pair with the wrapped native token (e.g. WETH) and the token to be accepted. The collector will then use the price of the token to determine if the fee is sufficient.

e.g. to find the Uniswap pool for the wrapped native token and USDC on arbitrum:

```sh
forge script --rpc-url $RPC_URL ./scripts/FindUniswapPool.s.sol --sig "run(address,address,address)" 0xf1D7CC64Fb4452F05c498126312eBE29f30Fbcf9 0x82af49447d8a07e3bd95bd0d56f35241523fbab1 0xaf88d065e77c8cc2239327c5edb3a432268e5831

== Logs ==
> 0xF64Dfe17C8b87F012FCf50FbDA1D62bfA148366a
```

Resulting config:

```toml
[collector]
  # ...

  [[collector.references]]
    token = "0xaf88d065e77c8cC2239327C5EDb3A432268e5831"

  [collector.references.uniswap_v2]
    pool = "0xF64Dfe17C8b87F012FCf50FbDA1D62bfA148366a"
    base_token = "0x82af49447d8a07e3bd95bd0d56f35241523fbab1"
```

## Register An Endorser

To register an endorser, the endorser must be deployed and the address must be added to the `TemporalRegistry` contract.

Deposit native tokens to register the endorser:

```sh
forge script --rpc-url $RPC_URL ./scripts/RegisterEndorser.s.sol --sig "lock(address,address,uint256)" <registry_address> <endorser_address> <lock_amount>
```

Remove registration by unlocking tokens:

```sh
forge script --rpc-url $RPC_URL ./scripts/RegisterEndorser.s.sol --sig "unlock(address,address)" <registry_address> <endorser_address>
```

Note: Unlocking is a 2 step process. The first run will initiate the unlock, the second run will finalize the unlock. Run the script twice to complete the unlock.
