// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.13;

import "../interfaces/Endorser.sol";


contract MockEndorser is Endorser {
  function isOperationReady(
    address,
    bytes calldata,
    bytes calldata _endorserCallData,
    uint256,
    uint256,
    uint256,
    address,
    uint256,
    uint256,
    bool
  ) external pure returns (
    bool readiness,
    GlobalDependency memory globalDependency,
    Dependency[] memory dependencies
  ) {
    (readiness, globalDependency, dependencies) = abi.decode(_endorserCallData, (bool, GlobalDependency, Dependency[]));
  }

  function encodeEndorserCalldata(
    bool _readiness,
    GlobalDependency memory _globalDependency,
    Dependency[] memory _dependencies
  ) external pure returns (bytes memory) {
    return abi.encode(_readiness, _globalDependency, _dependencies);
  }
}
