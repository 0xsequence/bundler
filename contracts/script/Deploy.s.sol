// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.18;

import {SingletonDeployer, console} from "erc2470-libs/script/SingletonDeployer.s.sol";
import {OperationValidator} from "../src/OperationValidator.sol";
import {TemporalRegistry} from "../src/TemporalRegistry.sol";

contract Deploy is SingletonDeployer {
    function run() external virtual {
        uint256 pk = vm.envUint("PRIVATE_KEY");
        bytes32 salt = bytes32(0);
        deployValidator(salt, pk);
        deployRegistry(salt, pk);
    }

    function deployValidator(bytes32 salt, uint256 pk) internal {
        bytes memory initCode = abi.encodePacked(
            type(OperationValidator).creationCode
        );
        _deployIfNotAlready("OperationValidator", initCode, salt, pk);
    }

    function deployRegistry(bytes32 salt, uint256 pk) internal {
        bytes memory initCode = abi.encodePacked(
            type(TemporalRegistry).creationCode
        );
        _deployIfNotAlready("TemporalRegistry", initCode, salt, pk);
    }
}
