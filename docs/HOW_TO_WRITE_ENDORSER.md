# How to write an ERC-5189 Endorser

Endorsers should primarily be written by wallet developers. The endorser is responsible for validating the operation and returning the required dependencies for the operation to be executed. To write an endorser, deep knowledge of the wallet's operation and storage space is required.

An endorser may not cover all possible edge cases, but it should cover the most common cases. The endorser should be efficient and provide only the necessary dependencies for the operation to be executed. When an operation is outside the endorser's scope, the endorser should return an invalid result.

Wallet interaction libraries should be updated to interact with the ERC-5189 protocol. This can be done invisibly to the user. Wallet developers are free to choose the endorser and operation parameters that best suits each transaction.

## Endorser Operation Checks

A "good" endorser will perform all necessary checks to ensure the operation is valid and provide only the essential dependencies for the operation to be executed. A "great" endorser will do this efficiently and with minimal computation costs.

The endorser may use `simulationSettings` to alter chain state prior to evaluation. This can allow an endorser to check storage slots that may not be accessible in the current state, or reduce computation needs by shortcutting known implementations.

Here is an incomplete list of checks that an endorser should perform.

### Valid Wallet Contract

The endorser should check that the operation's `entrypoint` address is a valid wallet contract.

If the wallet is a proxy, this will involve checking the bytecode at the `entrypoint` address to ensure it matches the expected wallet bytecode. The endorser must also list the expected implementation address as a constraint in the output dependencies.

### Valid Data

The endorser should check that the operation's `data` is a valid call to the wallet contract. This will involve decoding the `data`, checking the function signature and arguments.

### Valid Signature

Most smart contract wallets utilise a signature to validate a transaction. The endorser should check that the signature is valid and signed correctly.

The endorser should include storage slots accessed as dependencies or constraints when validating the signature. Consider `nonce` slots, `owner` slots, and `guard` slots as potential dependencies.

When validating the signature, the endorser should list the signers as constraints or dependencies in the output dependencies. In cases where the signature is signed by a contract, a nested approach should be taken.

If the operation `hasUntrustedContext`, the endorser may test signature validity within an `UntrustedContext`. Where possible, this should be avoided to reduce computation costs on the bundler and the risk of a dropped operation.

### Valid Repayment

Determining that the transaction is accurately repaid is essential. The endorser should calculate the gas costs and ensure that the operation repays the relayer for the execution.

The endorser must consider the operation gas parameters when determining repayment. As the transaction could be submitted with gas parameters that differ from those used in the transaction data, the endorser should include the gas parameters as dependencies where appropriate.

If repaying with an ERC-20, the endorser should include storage dependencies associated with the ERC-20 as well. This will ensure changes in wallet funds to not affect the operation's readiness state.

The endorser should consider the gas overhead cost of submitting the transaction, including calldata cost, as this will affect the repayment amount.

## Deployment

The endorser should be deployed to each network that supported wallets are deployed.

The endorser can then be registered with an endorser registry to establish a reputation. More information on endorser reputation can be found in the [ERC-5189 documentation on endorser registries](https://ercs.ethereum.org/ERCS/erc-5189#endorser-registry).

## What Happens If The Endorser Is Wrong?

If the endorser misreports readiness, the operation will be rejected by the bundler. The bundler will blacklist the endorser and flag an further operations with the endorser as invalid.

For more information on endorser reputation, please refer to the [ERC-5189 documentation on misbehavior detection](https://ercs.ethereum.org/ERCS/erc-5189#misbehavior-detection).
