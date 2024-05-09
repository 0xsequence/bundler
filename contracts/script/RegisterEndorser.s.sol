// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.18;

import {Script, console} from "forge-std/Script.sol";
import {TemporalRegistry} from "../src/TemporalRegistry.sol";

contract RegisterEndorser is Script {
    function lock(
        address temporalRegistry,
        address endorser,
        uint256 amount
    ) external {
        uint256 pk = vm.envUint("PRIVATE_KEY");
        address owner = vm.addr(pk);

        vm.broadcast(pk);
        TemporalRegistry(temporalRegistry).lock(owner, endorser, amount);
    }

    function unlock(address temporalRegistry, address endorser) external {
        uint256 pk = vm.envUint("PRIVATE_KEY");
        address owner = vm.addr(pk);

        uint256 idx = 0;
        while (true) {
            (
                uint40 unlockDuration,
                uint56 startedUnlock,
                address depositEndorser,
                ,
                address depositOwner
            ) = TemporalRegistry(temporalRegistry).deposits(idx);
            if (depositOwner == owner && depositEndorser == endorser) {
                console.log("Unlocking deposit %d", idx);
                if (startedUnlock == 0) {
                    // Start unlock
                    vm.broadcast(pk);
                    TemporalRegistry(temporalRegistry).startUnlock(idx);
                } else if (
                    unlockDuration >
                    block.timestamp - startedUnlock
                ) {
                    // Finish unlock
                    vm.broadcast(pk);
                    TemporalRegistry(temporalRegistry).unlock(idx);
                } else {
                    // Need to wait longer
                    console.log("Unlock duration has not passed yet");
                }
                break;
            }
            idx++;
        }
    }
}
