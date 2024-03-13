// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.18;


interface EndorserRegistry {
  event Burned(
      address indexed _endorser,
      address indexed _sender,
      uint256 _new,
      uint256 _total
  );

  function burn(address) external view returns (uint256);
}
