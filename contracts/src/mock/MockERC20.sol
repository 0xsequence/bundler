// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.18;

import "../interfaces/ERC20.sol";

// This isn't a true implementation of the ERC-20 spec. Use storage hacks to set balances.
contract MockERC20 is ERC20 {

  mapping(address => uint256) private balances;

  function balanceOf(address account) external view override returns (uint256) {
    return balances[account];
  }
}
