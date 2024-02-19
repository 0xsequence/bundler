// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.13;

import "./interfaces/ERC20.sol";
import "./interfaces/Endorser.sol";
import "./Math.sol";

contract OperationValidatorSimulator {
  error BundlerExecutionFailed();
  error BundlerUnderpaid(uint256 _paid, uint256 _expected);

  struct SimulationResult {
    bool paid;
    bool lied;
  }

  function fetchPaymentBal(address _feeToken) internal view returns (uint256) {
    if (_feeToken == address(0)) {
      return tx.origin.balance;
    } else {
      return ERC20(_feeToken).balanceOf(tx.origin);
    }
  }

  function simulateOperation(
    address _entrypoint,
    bytes calldata _data,
    bytes calldata _endorserCallData,
    uint256 _gasLimit,
    uint256 _maxFeePerGas,
    uint256 _maxPriorityFeePerGas,
    address _feeToken,
    uint256 _baseFeeScalingFactor,
    uint256 _baseFeeNormalizationFactor,
    bool _hasUntrustedContext,
    address _endorser,
    uint256 _calldataGas
  ) external view returns (SimulationResult memory result) {
  }

  function safeExecute(
    address _entrypoint,
    bytes calldata _data,
    uint256 _gasLimit,
    uint256 _maxFeePerGas,
    uint256 _maxPriorityFeePerGas,
    address _feeToken,
    uint256 _baseFeeScalingFactor,
    uint256 _baseFeeNormalizationFactor,
    uint256 _calldataGas
  ) external {
    uint256 preBal = fetchPaymentBal(_feeToken);
    
    uint256 preGas = gasleft();
    (bool ok,) = _entrypoint.call{ gas: _gasLimit }(_data);
    uint256 postGas = gasleft();

    if (!ok) {
      revert BundlerExecutionFailed();
    }

    uint256 postBal = fetchPaymentBal(_feeToken);

    uint256 gasUsed = preGas - postGas + _calldataGas;
    uint256 gasPrice = Math.min(Math.mulDiv(block.basefee, _baseFeeScalingFactor, _baseFeeNormalizationFactor) + _maxPriorityFeePerGas, _maxFeePerGas);
    uint256 expectPayment = gasUsed * gasPrice;
    uint256 paid = postBal - preBal;

    if (paid < expectPayment) {
      revert BundlerUnderpaid(paid, expectPayment);
    }
  }
}
