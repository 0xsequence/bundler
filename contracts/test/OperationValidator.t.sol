// SPDX-License-Identifier: Apache 2.0
pragma solidity ^0.8.13;

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
  function isOperationReady(
    address,
    bytes calldata,
    bytes calldata,
    uint256,
    uint256,
    uint256,
    address,
    uint256,
    uint256,
    bool
  ) external pure returns (
    bool readiness,
    GlobalDependency memory globalDependency,
    Dependency[] memory dependencies
  ) {}
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

    OperationValidator.SimulationResult memory result = 
      ov.simulateOperation(
        address(w),
        data,
        bytes(""),
        gasLimit,
        gasPrice,
        gasPrice,
        address(0),
        1,
        1,
        false,
        address(0)
      );
  
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

    OperationValidator.SimulationResult memory result = 
      ov.simulateOperation(
        address(w),
        data,
        bytes(""),
        gasLimit,
        gasPrice,
        gasPrice,
        address(t),
        1,
        1,
        false,
        address(0)
      );
  
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

    OperationValidator.SimulationResult memory result = 
      ov.simulateOperation(
        address(w),
        data,
        bytes(""),
        gasLimit,
        gasPrice,
        gasPrice,
        address(t),
        1,
        2,
        false,
        address(0)
      );
  
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

    OperationValidator.SimulationResult memory result = 
      ov.simulateOperation(
        address(w),
        data,
        bytes(""),
        gasLimit,
        gasPrice,
        gasPrice,
        address(0),
        1,
        1,
        false,
        address(e)
      );
  
    assertFalse(result.paid);
  }
}