// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.13;

import "./interfaces/ERC20.sol";
import "./interfaces/Endorser.sol";
import "./Math.sol";

contract OperationValidator {
  error BundlerExecutionFailed();
  error BundlerUnderpaid(uint256 _paid, uint256 _expected);

  struct SimulationResult {
    bool paid;
    bool readiness;
    Endorser.GlobalDependency globalDependency;
    Endorser.Dependency[] dependencies;
  }

  function fetchPaymentBal(address _feeToken) internal view returns (uint256) {
    if (_feeToken == address(0)) {
      return tx.origin.balance;
    } else {
      return ERC20(_feeToken).balanceOf(tx.origin);
    }
  }

  function _executeAndMeasureNoSideEffects(
    address _entrypoint,
    bytes calldata _data,
    uint256 _gasLimit,
    uint256 _maxFeePerGas,
    uint256 _maxPriorityFeePerGas,
    address _feeToken,
    uint256 _baseFeeScalingFactor,
    uint256 _baseFeeNormalizationFactor
  ) external returns (bool) {
    require(msg.sender == address(this), "only self");

    uint256 preBal = fetchPaymentBal(_feeToken);
    
    uint256 preGas = gasleft();
    // Ignore the return value, we don't trust any of it
    _entrypoint.call{ gas: _gasLimit }(_data);

    uint256 postGas = gasleft();
    uint256 postBal = fetchPaymentBal(_feeToken);

    uint256 gasUsed = preGas - postGas;
    uint256 baseFee = Math.mulDiv(block.basefee, _baseFeeScalingFactor, _baseFeeNormalizationFactor);
    uint256 gasPrice = Math.min(baseFee + _maxPriorityFeePerGas, _maxFeePerGas);
    uint256 expectPayment = gasUsed * gasPrice;

    if (postBal - preBal < expectPayment) {
      revert();
    }

    return true;
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
    address _endorser
  ) external returns (SimulationResult memory result) {
    // Try to execute the operation and measure the gas used
    // if it does not fail, then we can just return
    try OperationValidator(address(this))._executeAndMeasureNoSideEffects(
      _entrypoint,
      _data,
      _gasLimit,
      _maxFeePerGas,
      _maxPriorityFeePerGas,
      _feeToken,
      _baseFeeScalingFactor,
      _baseFeeNormalizationFactor
    ) returns (
      bool success
    ) {
      result.paid = success;
      return result;
    } catch {
      result.paid = false;
    }

    // We didn't got paid, we need to know
    // if the endorser considers the operation ready
    // if so, he lied to us. We need to use try-catch
    // as the endorser may revert instead of returning false.
    try Endorser(_endorser).isOperationReady(
      _entrypoint,
      _data,
      _endorserCallData,
      _gasLimit,
      _maxFeePerGas,
      _maxPriorityFeePerGas,
      _feeToken,
      _baseFeeScalingFactor,
      _baseFeeNormalizationFactor,
      _hasUntrustedContext
    ) returns (
      bool readiness,
      Endorser.GlobalDependency memory globalDependency,
      Endorser.Dependency[] memory dependencies
    ) {
      result.readiness = readiness;
      result.globalDependency = globalDependency;
      result.dependencies = dependencies;
      return result;
    } catch {
      result.readiness = false;
      return result;
    }
  }

  function safeExecute(
    address _entrypoint,
    bytes calldata _data,
    uint256 _gasLimit,
    uint256 _maxFeePerGas,
    uint256 _maxPriorityFeePerGas,
    address _feeToken,
    uint256 _baseFeeScalingFactor,
    uint256 _baseFeeNormalizationFactor
  ) external {
    uint256 preBal = fetchPaymentBal(_feeToken);
    
    uint256 preGas = gasleft();
    (bool ok,) = _entrypoint.call{ gas: _gasLimit }(_data);
    uint256 postGas = gasleft();

    if (!ok) {
      revert BundlerExecutionFailed();
    }

    uint256 postBal = fetchPaymentBal(_feeToken);

    uint256 gasUsed = preGas - postGas;
    uint256 gasPrice = Math.min(Math.mulDiv(block.basefee, _baseFeeScalingFactor, _baseFeeNormalizationFactor) + _maxPriorityFeePerGas, _maxFeePerGas);
    uint256 expectPayment = gasUsed * gasPrice;
    uint256 paid = postBal - preBal;

    if (paid < expectPayment) {
      revert BundlerUnderpaid(paid, expectPayment);
    }
  }
}
