// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.13;

import "solady/utils/FixedPointMathLib.sol";

import "./interfaces/Endorser.sol";
import "./interfaces/ERC20.sol";

contract OperationValidator {
  using FixedPointMathLib for *;

  error BundlerExecutionFailed();
  error BundlerUnderpaid(bool _succeed, uint256 _paid, uint256 _expected);

  struct SimulationResult {
    bool paid;
    bool readiness;
    Endorser.GlobalDependency globalDependency;
    Endorser.Dependency[] dependencies;
    bytes err;
  }

  function fetchPaymentBal(address _feeToken) internal view returns (uint256) {
    if (_feeToken == address(0)) {
      return tx.origin.balance;
    } else {
      return ERC20(_feeToken).balanceOf(tx.origin);
    }
  }

  function _executeAndMeasureNoSideEffects(Endorser.Operation calldata _op) external returns (bool) {
    require(msg.sender == address(this), "only self");

    uint256 preBal = fetchPaymentBal(_op.feeToken);
    
    uint256 preGas = gasleft();
    // Ignore the return value, we don't trust any of it
    (bool ok,) = _op.entrypoint.call{ gas: _op.gasLimit }(_op.data );

    uint256 postGas = gasleft();
    uint256 postBal = fetchPaymentBal(_op.feeToken);

    uint256 gasUsed = preGas - postGas;
    uint256 gasPrice = (block.basefee + _op.maxPriorityFeePerGas).min(_op.maxFeePerGas);
    uint256 expectPayment = (gasUsed * gasPrice).fullMulDiv(_op.feeScalingFactor, _op.feeNormalizationFactor);

    if (postBal - preBal < expectPayment) {
      revert BundlerUnderpaid(ok, postBal - preBal, expectPayment);
    }

    return true;
  }

  function simulateOperation(Endorser _endorser, Endorser.Operation calldata _op) external returns (SimulationResult memory result) {
    // Try to execute the operation and measure the gas used
    // if it does not fail, then we can just return
    try OperationValidator(address(this))._executeAndMeasureNoSideEffects(_op) returns (
      bool success
    ) {
      result.paid = success;
      return result;
    } catch (bytes memory err) {
      result.err = err;
      result.paid = false;
    }

    // We didn't got paid, we need to know
    // if the endorser considers the operation ready
    // if so, he lied to us. We need to use try-catch
    // as the endorser may revert instead of returning false.
    try _endorser.isOperationReady(_op) returns (
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
}
