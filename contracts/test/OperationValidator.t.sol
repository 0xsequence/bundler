// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.18;

import "forge-std/Test.sol";
import "solady/utils/SafeTransferLib.sol";
import "solady/tokens/ERC20.sol";

import { OperationValidator, Endorser } from "../src/OperationValidator.sol";

contract TestERC20 is ERC20 {
  function name() public override pure returns (string memory) {
    return "";
  }

  function symbol() public override pure returns (string memory) {
    return "";
  }

  function mint(address _to, uint256 _amount) external {
    _mint(_to, _amount);
  }

  function burn(address _from, uint256 _amount) external {
    _burn(_from, _amount);
  }
}

contract TestEndorser is Endorser {
  function isOperationReady(Endorser.Operation calldata _op) external pure returns (
    bool readiness,
    GlobalDependency memory globalDependency,
    Dependency[] memory dependencies
  ) {}

  function simulationSettings() external view returns (Endorser.Replacement[] memory replacements) {}
}

contract StubWallet {
  function payFee(address _token, uint256 _amount) external {
    if (_token == address(0)) {
      payable(address(tx.origin)).transfer(_amount);
    } else {
      SafeTransferLib.safeTransfer(_token, address(tx.origin), _amount);
    }
  }
  receive() external payable {}
}

contract OperationValidatorTest is Test {
  OperationValidator ov;

  function setUp() external {
    ov = new OperationValidator();
  }

  function testSimulateOperation() external {
    StubWallet w = new StubWallet();

    uint256 gasLimit = 10_000;
    uint256 gasPrice = 1 gwei;

    uint256 feePayment = gasLimit * gasPrice;

    vm.deal(address(w), feePayment);
    bytes memory data = abi.encodeWithSelector(
      w.payFee.selector,
      address(0),
      feePayment
    );

    Endorser.Operation memory op;
    op.entrypoint = address(w);
    op.data = data;
    op.gasLimit = gasLimit;
    op.maxFeePerGas = gasPrice;
    op.feeScalingFactor = 1;
    op.feeNormalizationFactor = 1;

    OperationValidator.SimulationResult memory result = ov.simulateOperation(Endorser(address(0)), op);
  
    assertTrue(result.paid);
  }

  function testSimulateOperationPayToken() external {
    TestERC20 t = new TestERC20();
    StubWallet w = new StubWallet();

    uint256 gasLimit = 24_000;
    uint256 gasPrice = 1 gwei;

    uint256 feePayment = gasLimit * gasPrice;

    t.mint(address(w), feePayment);
    bytes memory data = abi.encodeWithSelector(
      w.payFee.selector,
      address(t),
      feePayment
    );

    Endorser.Operation memory op;
    op.entrypoint = address(w);
    op.data = data;
    op.gasLimit = gasLimit;
    op.maxFeePerGas = gasPrice;
    op.maxPriorityFeePerGas = gasPrice;
    op.feeToken = address(t);
    op.feeScalingFactor = 1;
    op.feeNormalizationFactor = 1;

    OperationValidator.SimulationResult memory result = ov.simulateOperation(Endorser(address(0)), op);

    assertTrue(result.paid);
  }

  function testSimulateOperationPayTokenWithRate() external {
    TestERC20 t = new TestERC20();
    StubWallet w = new StubWallet();

    uint256 gasLimit = 24_000;
    uint256 gasPrice = 1 gwei;

    uint256 feePayment = gasLimit * gasPrice;

    t.mint(address(w), feePayment);
    bytes memory data = abi.encodeWithSelector(
      w.payFee.selector,
      address(t),
      feePayment / 2
    );

    Endorser.Operation memory op;
    op.entrypoint = address(w);
    op.data = data;
    op.gasLimit = gasLimit;
    op.maxFeePerGas = gasPrice;
    op.maxPriorityFeePerGas = gasPrice;
    op.feeToken = address(t);
    op.feeScalingFactor = 1;
    op.feeNormalizationFactor = 2;

    OperationValidator.SimulationResult memory result = ov.simulateOperation(Endorser(address(0)), op);

    assertTrue(result.paid);
  }

  function testSimulateOperationUnderpays() external {
    StubWallet w = new StubWallet();

    uint256 gasLimit = 10_000;
    uint256 gasPrice = 1 gwei;

    uint256 feePayment = gasLimit * gasPrice;

    vm.deal(address(w), feePayment);
    bytes memory data = abi.encodeWithSelector(
      w.payFee.selector,
      address(0),
      feePayment / 2
    );

    TestEndorser e = new TestEndorser();

    Endorser.Operation memory op;
    op.entrypoint = address(w);
    op.data = data;
    op.gasLimit = gasLimit;
    op.maxFeePerGas = gasPrice;
    op.maxPriorityFeePerGas = gasPrice;
    op.feeScalingFactor = 1;
    op.feeNormalizationFactor = 1;

    OperationValidator.SimulationResult memory result = ov.simulateOperation(Endorser(address(e)), op);
  
    assertFalse(result.paid);
  }
}