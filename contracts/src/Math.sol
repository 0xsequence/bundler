// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.13;

library Math {
  function max(uint256 a, uint256 b) internal pure returns (uint256) {
    return a > b ? a : b;
  }

  function min(uint256 a, uint256 b) internal pure returns (uint256) {
    return a < b ? a : b;
  }
}
