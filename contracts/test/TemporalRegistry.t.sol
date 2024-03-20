// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.18;

import "forge-std/Test.sol";

import { TemporalRegistry } from "../src/TemporalRegistry.sol";


contract TemporalRegistryTest is Test {
  TemporalRegistry registry;

  function assumeNoSys(address _a) internal pure {
    vm.assume(_a != address(0x007109709ecfa91a80626ff3989d68f67f5b1dd12d));
    vm.assume(_a != address(0x004e59b44847b379578588920ca78fbf26c0b4956c));
    vm.assume(_a != address(0x00000000000000000000636f6e736f6c652e6c6f67));
    vm.assume(uint160(_a)> 10);
  }
  function setUp() external {
    registry = new TemporalRegistry();
  }

  function testLock(
    address _from,
    address _owner,
    address _endorser,
    uint256 _amount,
    uint256 _unlockDuration
  ) external {
    assumeNoSys(_from);
    _amount = bound(_amount, 0, 100000 ether);
    vm.assume(_unlockDuration != 0);

    vm.deal(_from, _amount);
    vm.prank(_from);
    registry.lock{ value: _amount }(_owner, _endorser, _unlockDuration);

    (
      uint40 unlockTime,
      uint56 startedUnlock,
      address endorser,
      uint96 amount,
      address owner
    ) = registry.deposits(0);

    uint256 maxTime = registry.MAX_UNLOCK();
    assertEq(unlockTime, uint40(_unlockDuration > maxTime ? maxTime : _unlockDuration));
    assertEq(startedUnlock, uint56(0));
    assertEq(endorser, _endorser);
    assertEq(amount, uint96(_amount));
    assertEq(owner, _owner);

    uint256 weight = registry.weightFor(_unlockDuration, _amount);
    assertEq(registry.burn(_endorser), weight);

    assertEq(address(registry).balance, _amount);
  }

  function testLockTwo(
    address _from1,
    address _from2,
    address _endorser,
    uint256 _amount1,
    uint256 _amount2,
    uint256 _unlockDuration1,
    uint256 _unlockDuration2
  ) external {
    assumeNoSys(_from1);
    assumeNoSys(_from2);
    _amount1 = bound(_amount1, 0, 100000 ether);
    _amount2 = bound(_amount2, 0, 100000 ether);
    vm.assume(_unlockDuration1 != 0);
    vm.assume(_unlockDuration2 != 0);


    if (_from1 != _from2) {
      vm.deal(_from1, _amount1);
      vm.deal(_from2, _amount2);
    } else {
      vm.deal(_from1, _amount1 + _amount2);
    }

    vm.prank(_from1);
    registry.lock{ value: _amount1 }(_from1, _endorser, _unlockDuration1);
    vm.prank(_from2);
    registry.lock{ value: _amount2 }(_from2, _endorser, _unlockDuration2);

    uint256 maxTime = registry.MAX_UNLOCK();

    {
      (
        uint40 unlockTime1,
        uint56 startedUnlock1,
        address endorser1,
        uint96 amount1,
        address owner1
      ) = registry.deposits(0);

      assertEq(unlockTime1, uint40(_unlockDuration1 > maxTime ? maxTime : _unlockDuration1));
      assertEq(startedUnlock1, uint56(0));
      assertEq(endorser1, _endorser);
      assertEq(amount1, uint96(_amount1));
      assertEq(owner1, _from1);
    }

    {
      (
        uint40 unlockTime2,
        uint56 startedUnlock2,
        address endorser2,
        uint96 amount2,
        address owner2
      ) = registry.deposits(1);

      assertEq(unlockTime2, uint40(_unlockDuration2 > maxTime ? maxTime : _unlockDuration2));
      assertEq(startedUnlock2, uint56(0));
      assertEq(endorser2, _endorser);
      assertEq(amount2, uint96(_amount2));
      assertEq(owner2, _from2);
    }

    uint256 weight1 = registry.weightFor(_unlockDuration1, _amount1);
    uint256 weight2 = registry.weightFor(_unlockDuration2, _amount2);
    assertEq(registry.burn(_endorser), weight1 + weight2);

    assertEq(address(registry).balance, _amount1 + _amount2);
  }

  function testStartUnlock(
    address _from,
    address _owner,
    address _endorser,
    uint256 _amount,
    uint256 _unlockDuration,
    uint256 _blockTime
  ) external {
    assumeNoSys(_from);
    _amount = bound(_amount, 0, 100000 ether);
    _unlockDuration = bound(_unlockDuration, 1, type(uint64).max - 1);

    vm.deal(_from, _amount);
    vm.prank(_from);
    registry.lock{ value: _amount }(_owner, _endorser, _unlockDuration);

    vm.warp(_blockTime);
    vm.prank(_owner);
    registry.startUnlock(0);

    (
      uint40 unlockTime,
      uint56 startedUnlock,
      address endorser,
      uint96 amount,
      address owner
    ) = registry.deposits(0);


    uint256 maxTime = registry.MAX_UNLOCK();
    assertEq(unlockTime, uint40(_unlockDuration > maxTime ? maxTime : _unlockDuration));
    assertEq(startedUnlock, uint56(block.timestamp));
    assertEq(endorser, _endorser);
    assertEq(amount, uint96(_amount));
    assertEq(owner, _owner);

    assertEq(registry.burn(_endorser), 0);
    assertEq(address(registry).balance, _amount);
  }

  function testStartUnlockTwice(
    address _from1,
    address _from2,
    address _owner1,
    address _owner2,
    address _endorser,
    uint256 _amount1,
    uint256 _amount2,
    uint256 _unlockDuration1,
    uint256 _unlockDuration2,
    uint256 _blockTime
  ) external {
    assumeNoSys(_from1);
    assumeNoSys(_from2);
    _amount1 = bound(_amount1, 0, 100000 ether);
    _amount2 = bound(_amount2, 0, 100000 ether);
    _unlockDuration1 = bound(_unlockDuration1, 1, type(uint64).max - 1);
    _unlockDuration2 = bound(_unlockDuration2, 1, type(uint64).max - 1);

    if (_from1 != _from2) {
      vm.deal(_from1, _amount1);
      vm.deal(_from2, _amount2);
    } else {
      vm.deal(_from1, _amount1 + _amount2);
    }

    vm.prank(_from1);
    registry.lock{ value: _amount1 }(_owner1, _endorser, _unlockDuration1);
    vm.prank(_from2);
    registry.lock{ value: _amount2 }(_owner2, _endorser, _unlockDuration2);

    vm.warp(_blockTime);
    vm.prank(_owner1);
    registry.startUnlock(0);

    uint256 maxTime = registry.MAX_UNLOCK();

    {
      (
        uint40 unlockTime1,
        uint56 startedUnlock1,
        address endorser1,
        uint96 amount1,
        address owner1
      ) = registry.deposits(0);

      assertEq(unlockTime1, uint40(_unlockDuration1 > maxTime ? maxTime : _unlockDuration1));
      assertEq(startedUnlock1, uint56(_blockTime));
      assertEq(endorser1, _endorser);
      assertEq(amount1, uint96(_amount1));
      assertEq(owner1, _owner1);
    }

    {
      (
        uint40 unlockTime2,
        uint56 startedUnlock2,
        address endorser2,
        uint96 amount2,
        address owner2
      ) = registry.deposits(1);

      assertEq(unlockTime2, uint40(_unlockDuration2 > maxTime ? maxTime : _unlockDuration2));
      assertEq(startedUnlock2, uint56(0));
      assertEq(endorser2, _endorser);
      assertEq(amount2, uint96(_amount2));
      assertEq(owner2, _owner2);
    }

    uint256 weight2 = registry.weightFor(_unlockDuration2, _amount2);
    assertEq(registry.burn(_endorser), weight2);
    assertEq(address(registry).balance, _amount1 + _amount2);
  }

  function testUnlock(
    address _anyCaller,
    address _from,
    address _owner,
    address _endorser,
    uint256 _amount,
    uint256 _unlockDuration,
    uint256 _blockTime,
    uint256 _waitTime
  ) external {
    assumeNoSys(_from);
    _amount = bound(_amount, 0, 100000 ether);
    _unlockDuration = bound(_unlockDuration, 1, type(uint64).max - 1);
    _blockTime = bound(_blockTime, _unlockDuration, type(uint64).max);
    _waitTime = bound(_waitTime, _unlockDuration + 1, type(uint64).max);

    vm.deal(_from, _amount);
    vm.prank(_from);
    registry.lock{ value: _amount }(_owner, _endorser, _unlockDuration);

    vm.warp(_blockTime);
    vm.prank(_owner);
    registry.startUnlock(0);

    vm.warp(_waitTime + _blockTime);
    vm.prank(_anyCaller);
    registry.unlock(0);

    (
      uint40 unlockTime,
      uint56 startedUnlock,
      address endorser,
      uint96 amount,
      address owner
    ) = registry.deposits(0);

    assertEq(unlockTime, 0);
    assertEq(startedUnlock, 0);
    assertEq(endorser, address(0));
    assertEq(amount, 0);
    assertEq(owner, address(0));

    assertEq(registry.burn(_endorser), 0);
    assertEq(address(registry).balance, 0);
  }
}
