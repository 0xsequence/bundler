// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.18;

contract MockWallet {
  function execute(address payable[] memory _to, uint256[] memory _value) external {
    for (uint256 i = 0; i < _to.length; i++) {
      payable(_to[i]).transfer(_value[i]);
    }
  }

  receive() external payable {}
}
