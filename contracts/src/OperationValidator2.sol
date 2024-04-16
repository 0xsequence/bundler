// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.18;

import "./interfaces/Endorser.sol";
import "./interfaces/ERC20.sol";


contract OperationValidator2 {
  struct SimulationResult2 {
    uint256 payment;
    uint256 gasUsed;
  }

  function fetchPaymentBal(address _feeToken) internal view returns (uint256) {
    if (_feeToken == address(0)) {
      return tx.origin.balance;
    } else {
      return ERC20(_feeToken).balanceOf(tx.origin);
    }
  }

  function simulateOperation(Endorser.Operation calldata _op) external returns (SimulationResult2 memory result) {
    
    uint256 preBal = fetchPaymentBal(_op.feeToken);
    uint256 preGas = gasleft();

    _op.entrypoint.call{ gas: _op.gasLimit }(_op.data );
    
    uint256 postGas = gasleft();
    uint256 postBal = fetchPaymentBal(_op.feeToken);

    result.payment = postBal - preBal;
    result.gasUsed = preGas - postGas;

    return result;
  }
}
