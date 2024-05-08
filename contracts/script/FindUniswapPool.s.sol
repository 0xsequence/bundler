// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.18;

import {Script, console} from "forge-std/Script.sol";

interface IUniswapV2Factory {
    function getPair(address tokenA, address tokenB) external view returns (address pair);
}

/**
 * This contract is used to find a Uniswap pool for a given pair of tokens.
 * The Uniswap V2 Factory address can be found here: https://docs.uniswap.org/contracts/v2/reference/smart-contracts/v2-deployments
 */
contract FindUniswapPool is Script {
    function run(
        address v2Factory,
        address wrappedNativeToken,
        address erc20Token
    ) external virtual {
        (address token0, address token1) = wrappedNativeToken < erc20Token
            ? (wrappedNativeToken, erc20Token)
            : (erc20Token, wrappedNativeToken);
        address pair = IUniswapV2Factory(v2Factory).getPair(token0, token1);
        console.log("Uniswap pool for %s and %s: %s", token0, token1, pair);
    }
}
