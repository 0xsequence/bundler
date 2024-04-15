// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.18;

import "forge-std/Test.sol";
import { HuffDeployer } from "foundry-huff/HuffDeployer.sol";


contract BatchCallerTest is Test {
  address caller;

  function setUp() external {
    caller = HuffDeployer
      .config()
      .with_evm_version("paris")
      .deploy("tools/BatchCaller");
  }

  struct Call {
    bytes callData;
    bytes returnData;
  }

  function testSimpleCall() external {
    bytes memory callData = abi.encodeWithSignature("test()");
    bytes memory returnData = abi.encodeWithSignature("testReturn()");
    uint256 biggestCalldata = callData.length;

    address to = address(0x999999cf1046e68e36E1aA2E0E07105eDDD1f08E);

    vm.mockCall(to, callData, returnData);

    bytes memory batchData = abi.encodePacked(
      biggestCalldata,
      abi.encode(to),
      uint256(callData.length),
      callData
    );

    (bool ok, bytes memory res) = caller.call{ gas: 200_000 }(batchData);
    assertTrue(ok);

    bytes memory expected = abi.encodePacked(
      uint256(4),
      returnData
    );

    assertEq(res, expected);
  }

  function testManyCalls(Call[] calldata _calls) external {
    unchecked {
      uint256 highestSize;
      for (uint256 i = 0; i < _calls.length; i++) {
        if (_calls[i].callData.length > highestSize) {
          highestSize = _calls[i].callData.length;
        }
      }

      bytes memory callData = abi.encodePacked(
        highestSize
      );

      bytes memory expected = hex"";

      for (uint256 i = 0; i < _calls.length; i++) {
        // Generate random address
        address addr = address(uint160(uint256(keccak256(abi.encodePacked(i)))));
        vm.mockCall(addr, _calls[i].callData, _calls[i].returnData);

        callData = abi.encodePacked(
          callData,
          abi.encode(addr),
          uint256(_calls[i].callData.length),
          _calls[i].callData
        );

        expected = abi.encodePacked(
          expected,
          uint256(_calls[i].returnData.length),
          _calls[i].returnData
        );
      }

      (bool ok, bytes memory res) = caller.call{ gas: 200_000 }(callData);
      assertTrue(ok);
      assertEq(res, expected);
    }
  }
}