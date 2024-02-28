// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.13;


interface Endorser {
  struct GlobalDependency {
    bool basefee;
    bool blobbasefee;
    bool chainid;
    bool coinbase;
    bool difficulty;
    bool gasLimit;
    bool number;
    bool timestamp;
    bool txOrigin;
    bool txGasPrice;
    uint256 maxBlockNumber;
    uint256 maxBlockTimestamp;
  }

  struct Constraint {
    bytes32 slot;
    bytes32 minValue;
    bytes32 maxValue;
  }

  struct Dependency {
    address addr;
    bool balance;
    bool code;
    bool nonce;
    bool allSlots;
    bytes32[] slots;
    Constraint[] constraints;
  }

  function isOperationReady(
    address _entrypoint,
    bytes calldata _data,
    bytes calldata _endorserCallData,
    uint256 _gasLimit,
    uint256 _maxFeePerGas,
    uint256 _maxPriorityFeePerGas,
    address _feeToken,
    uint256 _baseFeeScalingFactor,
    uint256 _baseFeeNormalizationFactor,
    bool _hasUntrustedContext
  ) external returns (
    bool readiness,
    GlobalDependency memory globalDependency,
    Dependency[] memory dependencies
  );
}
