// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.18;

import "../interfaces/Endorser.sol";


contract MockEndorser is Endorser {
  function isOperationReady(Endorser.Operation calldata _op) external pure returns (
    bool readiness,
    GlobalDependency memory globalDependency,
    Dependency[] memory dependencies
  ) {
    (readiness, globalDependency, dependencies) = abi.decode(_op.endorserCallData, (bool, GlobalDependency, Dependency[]));
  }

  function encodeEndorserCalldata(
    bool _readiness,
    GlobalDependency memory _globalDependency,
    Dependency[] memory _dependencies
  ) external pure returns (bytes memory) {
    return abi.encode(_readiness, _globalDependency, _dependencies);
  }

  function simulationSettings(Endorser.Operation calldata) external pure returns (Replacement[] memory replacements) {
    return new Replacement[](0);
  }
}
