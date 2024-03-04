// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.13;

import "forge-std/Test.sol";

import "../src/OperationValidator.sol";

contract StubWallet {
  function payFee(address _toke, uint256 _amount) external {
    payable(address(tx.origin)).transfer(_amount);
  }
}

contract OperationValidatorTest is Test {
  OperationValidator ov;

  function setUp() external {
    ov = new OperationValidator();
  }

  function testSimulateOperation() external {

  }
}