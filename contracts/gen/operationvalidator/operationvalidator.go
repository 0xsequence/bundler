// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package operationvalidator

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
	_ = abi.ConvertType
)

// OperationValidatorSimulationResult is an auto generated low-level Go binding around an user-defined struct.
type OperationValidatorSimulationResult struct {
	Paid bool
	Lied bool
}

// OperationvalidatorMetaData contains all meta data concerning the Operationvalidator contract.
var OperationvalidatorMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"safeExecute\",\"inputs\":[{\"name\":\"_entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_baseFeeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_baseFeeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_calldataGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"simulateOperation\",\"inputs\":[{\"name\":\"_entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_endorserCallData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_baseFeeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_baseFeeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_hasUntrustedContext\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"_endorser\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_calldataGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"result\",\"type\":\"tuple\",\"internalType\":\"structOperationValidator.SimulationResult\",\"components\":[{\"name\":\"paid\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"lied\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"nonpayable\"},{\"type\":\"error\",\"name\":\"BundlerExecutionFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"BundlerUnderpaid\",\"inputs\":[{\"name\":\"_paid\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_expected\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"DIVISION_BY_ZERO\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UNDER_OVERFLOW\",\"inputs\":[]}]",
}

// OperationvalidatorABI is the input ABI used to generate the binding from.
// Deprecated: Use OperationvalidatorMetaData.ABI instead.
var OperationvalidatorABI = OperationvalidatorMetaData.ABI

// Operationvalidator is an auto generated Go binding around an Ethereum contract.
type Operationvalidator struct {
	OperationvalidatorCaller     // Read-only binding to the contract
	OperationvalidatorTransactor // Write-only binding to the contract
	OperationvalidatorFilterer   // Log filterer for contract events
}

// OperationvalidatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type OperationvalidatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationvalidatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OperationvalidatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationvalidatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OperationvalidatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationvalidatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OperationvalidatorSession struct {
	Contract     *Operationvalidator // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// OperationvalidatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OperationvalidatorCallerSession struct {
	Contract *OperationvalidatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// OperationvalidatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OperationvalidatorTransactorSession struct {
	Contract     *OperationvalidatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// OperationvalidatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type OperationvalidatorRaw struct {
	Contract *Operationvalidator // Generic contract binding to access the raw methods on
}

// OperationvalidatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OperationvalidatorCallerRaw struct {
	Contract *OperationvalidatorCaller // Generic read-only contract binding to access the raw methods on
}

// OperationvalidatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OperationvalidatorTransactorRaw struct {
	Contract *OperationvalidatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOperationvalidator creates a new instance of Operationvalidator, bound to a specific deployed contract.
func NewOperationvalidator(address common.Address, backend bind.ContractBackend) (*Operationvalidator, error) {
	contract, err := bindOperationvalidator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Operationvalidator{OperationvalidatorCaller: OperationvalidatorCaller{contract: contract}, OperationvalidatorTransactor: OperationvalidatorTransactor{contract: contract}, OperationvalidatorFilterer: OperationvalidatorFilterer{contract: contract}}, nil
}

// NewOperationvalidatorCaller creates a new read-only instance of Operationvalidator, bound to a specific deployed contract.
func NewOperationvalidatorCaller(address common.Address, caller bind.ContractCaller) (*OperationvalidatorCaller, error) {
	contract, err := bindOperationvalidator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OperationvalidatorCaller{contract: contract}, nil
}

// NewOperationvalidatorTransactor creates a new write-only instance of Operationvalidator, bound to a specific deployed contract.
func NewOperationvalidatorTransactor(address common.Address, transactor bind.ContractTransactor) (*OperationvalidatorTransactor, error) {
	contract, err := bindOperationvalidator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OperationvalidatorTransactor{contract: contract}, nil
}

// NewOperationvalidatorFilterer creates a new log filterer instance of Operationvalidator, bound to a specific deployed contract.
func NewOperationvalidatorFilterer(address common.Address, filterer bind.ContractFilterer) (*OperationvalidatorFilterer, error) {
	contract, err := bindOperationvalidator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OperationvalidatorFilterer{contract: contract}, nil
}

// bindOperationvalidator binds a generic wrapper to an already deployed contract.
func bindOperationvalidator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OperationvalidatorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Operationvalidator *OperationvalidatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Operationvalidator.Contract.OperationvalidatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Operationvalidator *OperationvalidatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Operationvalidator.Contract.OperationvalidatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Operationvalidator *OperationvalidatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Operationvalidator.Contract.OperationvalidatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Operationvalidator *OperationvalidatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Operationvalidator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Operationvalidator *OperationvalidatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Operationvalidator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Operationvalidator *OperationvalidatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Operationvalidator.Contract.contract.Transact(opts, method, params...)
}

// SafeExecute is a paid mutator transaction binding the contract method 0x3e505d4d.
//
// Solidity: function safeExecute(address _entrypoint, bytes _data, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, uint256 _calldataGas) returns()
func (_Operationvalidator *OperationvalidatorTransactor) SafeExecute(opts *bind.TransactOpts, _entrypoint common.Address, _data []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Operationvalidator.contract.Transact(opts, "safeExecute", _entrypoint, _data, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _calldataGas)
}

// SafeExecute is a paid mutator transaction binding the contract method 0x3e505d4d.
//
// Solidity: function safeExecute(address _entrypoint, bytes _data, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, uint256 _calldataGas) returns()
func (_Operationvalidator *OperationvalidatorSession) SafeExecute(_entrypoint common.Address, _data []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Operationvalidator.Contract.SafeExecute(&_Operationvalidator.TransactOpts, _entrypoint, _data, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _calldataGas)
}

// SafeExecute is a paid mutator transaction binding the contract method 0x3e505d4d.
//
// Solidity: function safeExecute(address _entrypoint, bytes _data, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, uint256 _calldataGas) returns()
func (_Operationvalidator *OperationvalidatorTransactorSession) SafeExecute(_entrypoint common.Address, _data []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Operationvalidator.Contract.SafeExecute(&_Operationvalidator.TransactOpts, _entrypoint, _data, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _calldataGas)
}

// SimulateOperation is a paid mutator transaction binding the contract method 0x052e7f32.
//
// Solidity: function simulateOperation(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext, address _endorser, uint256 _calldataGas) returns((bool,bool) result)
func (_Operationvalidator *OperationvalidatorTransactor) SimulateOperation(opts *bind.TransactOpts, _entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool, _endorser common.Address, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Operationvalidator.contract.Transact(opts, "simulateOperation", _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext, _endorser, _calldataGas)
}

// SimulateOperation is a paid mutator transaction binding the contract method 0x052e7f32.
//
// Solidity: function simulateOperation(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext, address _endorser, uint256 _calldataGas) returns((bool,bool) result)
func (_Operationvalidator *OperationvalidatorSession) SimulateOperation(_entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool, _endorser common.Address, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Operationvalidator.Contract.SimulateOperation(&_Operationvalidator.TransactOpts, _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext, _endorser, _calldataGas)
}

// SimulateOperation is a paid mutator transaction binding the contract method 0x052e7f32.
//
// Solidity: function simulateOperation(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext, address _endorser, uint256 _calldataGas) returns((bool,bool) result)
func (_Operationvalidator *OperationvalidatorTransactorSession) SimulateOperation(_entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool, _endorser common.Address, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Operationvalidator.Contract.SimulateOperation(&_Operationvalidator.TransactOpts, _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext, _endorser, _calldataGas)
}
