// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.18;

import "./interfaces/Registry.sol";
import "solady/utils/FixedPointMathLib.sol";


contract TemporalRegistry is EndorserRegistry {
  using FixedPointMathLib for *;

  error NotOwner();
  error UnlockStarted();
  error NoUnlockStarted();
  error NotUnlocked();
  error RecoverFailed();
  error UnlockZero();

  event UnBurned(
    address indexed _endorser,
    address indexed _sender,
    uint256 _new,
    uint256 _total
  );

  event Locked(
    address indexed _owner,
    address indexed _endorser,
    uint256 _amount,
    uint256 _unlockDuration,
    uint256 _index
  );

  event Recovered(
    address indexed _owner,
    address indexed _endorser,
    uint256 _amount
  );

  struct Deposit {
    uint40 unlockDuration;
    uint56 startedUnlock;
    address endorser;
    uint96 amount;
    address owner;
  }

  uint256 public constant MAX_UNLOCK = 365 days * 10;

  Deposit[] public deposits;
  mapping(address => uint256) public burn;

  function weightFor(
    uint256 _time,
    uint256 _amount
  ) public pure returns (uint256) {
    uint256 ctime = _time.min(MAX_UNLOCK);
    return _amount.fullMulDiv(ctime, MAX_UNLOCK);
  }

  function lock(
    address _owner,
    address _endorser,
    uint256 _unlockDuration
  ) external payable {
    if (_unlockDuration == 0) {
      revert UnlockZero();
    }

    _unlockDuration = _unlockDuration.min(MAX_UNLOCK);

    deposits.push(
      Deposit({
        unlockDuration: uint40(_unlockDuration),
        startedUnlock: 0,
        endorser: _endorser,
        amount: uint96(msg.value),
        owner: _owner
      })
    );

    emit Locked(_owner, _endorser, msg.value, _unlockDuration, deposits.length - 1);

    uint256 weight = weightFor(_unlockDuration, msg.value);
    burn[_endorser] += weight;
    emit Burned(_endorser, _owner, weight, burn[_endorser]);
  }

  function startUnlock(
    uint256 _index
  ) external {
    Deposit storage deposit = deposits[_index];
    if (deposit.owner != msg.sender) {
      revert NotOwner();
    }

    if (deposit.startedUnlock != 0) {
      revert UnlockStarted();
    }

    deposit.startedUnlock = uint56(block.timestamp);

    uint256 weight = weightFor(deposit.unlockDuration, deposit.amount);
    burn[deposit.endorser] -= weight;
    emit UnBurned(deposit.endorser, msg.sender, weight, burn[deposit.endorser]);
  }

  function unlock(
    uint256 _index
  ) external {
    Deposit storage deposit = deposits[_index];
    if (deposit.startedUnlock == 0) {
      revert NoUnlockStarted();
    }

    if (block.timestamp < uint256(deposit.startedUnlock) + uint256(deposit.unlockDuration)) {
      revert NotUnlocked();
    }

    emit Recovered(deposit.owner, deposit.endorser, deposit.amount);

    uint256 amount = deposit.amount;
    delete deposits[_index];
    (bool ok,) = payable(deposit.owner).call{ value: amount }("");
    if (!ok) {
      revert RecoverFailed();
    }
  }
}
