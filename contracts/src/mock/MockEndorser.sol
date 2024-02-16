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

  function randomEndorserCalldata(bool _ready, uint256 _seed) external view returns (bytes memory) {
    uint256 dependenciesCount = _randomNum(_seed) % 10;
    Dependency[] memory dependencies = new Dependency[](dependenciesCount);
    for (uint256 i = 0; i < dependenciesCount; i++) {
      uint256 slotsCount = _randomNum(_seed) % 10;
      bytes32[] memory slots = new bytes32[](slotsCount);

      for (uint256 j = 0; j < slotsCount; j++) {
        slots[j] = bytes32(_randomNum(_seed));
      }

      dependencies[i] = Dependency({
        addr: address(0),
        balance: _randomNum(_seed + i) % 2 == 0,
        code: _randomNum(_seed + i + 1) % 2 == 0,
        nonce: _randomNum(_seed + i + 2) % 2 == 0,
        allSlots: _randomNum(_seed + i + 3) % 2 == 0,
        slots: slots,
        constraints: new Constraint[](0)
      });
    }

    GlobalDependency memory gd;
    return abi.encode(_ready, gd, dependencies);
  }

  function _randomNum(uint256 _seed) internal view returns (uint256) {
    return uint256(keccak256(abi.encodePacked(_seed, gasleft())));
  }
}
