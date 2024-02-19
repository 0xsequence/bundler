// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package operationvalidator

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/0xsequence/ethkit/go-ethereum"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi"
	"github.com/0xsequence/ethkit/go-ethereum/accounts/abi/bind"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/core/types"
	"github.com/0xsequence/ethkit/go-ethereum/event"
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

// OperationValidatorSimulatorSimulationResult is an auto generated low-level Go binding around an user-defined struct.
type OperationValidatorSimulatorSimulationResult struct {
	Paid bool
	Lied bool
}

// OperationValidatorSimulatorMetaData contains all meta data concerning the OperationValidatorSimulator contract.
var OperationValidatorSimulatorMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"safeExecute\",\"inputs\":[{\"name\":\"_entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_baseFeeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_baseFeeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_calldataGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"simulateOperation\",\"inputs\":[{\"name\":\"_entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_endorserCallData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_baseFeeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_baseFeeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_hasUntrustedContext\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"_endorser\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_calldataGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"result\",\"type\":\"tuple\",\"internalType\":\"structOperationValidatorSimulator.SimulationResult\",\"components\":[{\"name\":\"paid\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"lied\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"view\"},{\"type\":\"error\",\"name\":\"BundlerExecutionFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"BundlerUnderpaid\",\"inputs\":[{\"name\":\"_paid\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_expected\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"DIVISION_BY_ZERO\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UNDER_OVERFLOW\",\"inputs\":[]}]",
}

// OperationValidatorSimulatorABI is the input ABI used to generate the binding from.
// Deprecated: Use OperationValidatorSimulatorMetaData.ABI instead.
var OperationValidatorSimulatorABI = OperationValidatorSimulatorMetaData.ABI

// OperationValidatorSimulator is an auto generated Go binding around an Ethereum contract.
type OperationValidatorSimulator struct {
	OperationValidatorSimulatorCaller     // Read-only binding to the contract
	OperationValidatorSimulatorTransactor // Write-only binding to the contract
	OperationValidatorSimulatorFilterer   // Log filterer for contract events
}

// OperationValidatorSimulatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type OperationValidatorSimulatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationValidatorSimulatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OperationValidatorSimulatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationValidatorSimulatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OperationValidatorSimulatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationValidatorSimulatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OperationValidatorSimulatorSession struct {
	Contract     *OperationValidatorSimulator // Generic contract binding to set the session for
	CallOpts     bind.CallOpts                // Call options to use throughout this session
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// OperationValidatorSimulatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OperationValidatorSimulatorCallerSession struct {
	Contract *OperationValidatorSimulatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts                      // Call options to use throughout this session
}

// OperationValidatorSimulatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OperationValidatorSimulatorTransactorSession struct {
	Contract     *OperationValidatorSimulatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts                      // Transaction auth options to use throughout this session
}

// OperationValidatorSimulatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type OperationValidatorSimulatorRaw struct {
	Contract *OperationValidatorSimulator // Generic contract binding to access the raw methods on
}

// OperationValidatorSimulatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OperationValidatorSimulatorCallerRaw struct {
	Contract *OperationValidatorSimulatorCaller // Generic read-only contract binding to access the raw methods on
}

// OperationValidatorSimulatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OperationValidatorSimulatorTransactorRaw struct {
	Contract *OperationValidatorSimulatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOperationValidatorSimulator creates a new instance of OperationValidatorSimulator, bound to a specific deployed contract.
func NewOperationValidatorSimulator(address common.Address, backend bind.ContractBackend) (*OperationValidatorSimulator, error) {
	contract, err := bindOperationValidatorSimulator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OperationValidatorSimulator{OperationValidatorSimulatorCaller: OperationValidatorSimulatorCaller{contract: contract}, OperationValidatorSimulatorTransactor: OperationValidatorSimulatorTransactor{contract: contract}, OperationValidatorSimulatorFilterer: OperationValidatorSimulatorFilterer{contract: contract}}, nil
}

// NewOperationValidatorSimulatorCaller creates a new read-only instance of OperationValidatorSimulator, bound to a specific deployed contract.
func NewOperationValidatorSimulatorCaller(address common.Address, caller bind.ContractCaller) (*OperationValidatorSimulatorCaller, error) {
	contract, err := bindOperationValidatorSimulator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OperationValidatorSimulatorCaller{contract: contract}, nil
}

// NewOperationValidatorSimulatorTransactor creates a new write-only instance of OperationValidatorSimulator, bound to a specific deployed contract.
func NewOperationValidatorSimulatorTransactor(address common.Address, transactor bind.ContractTransactor) (*OperationValidatorSimulatorTransactor, error) {
	contract, err := bindOperationValidatorSimulator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OperationValidatorSimulatorTransactor{contract: contract}, nil
}

// NewOperationValidatorSimulatorFilterer creates a new log filterer instance of OperationValidatorSimulator, bound to a specific deployed contract.
func NewOperationValidatorSimulatorFilterer(address common.Address, filterer bind.ContractFilterer) (*OperationValidatorSimulatorFilterer, error) {
	contract, err := bindOperationValidatorSimulator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OperationValidatorSimulatorFilterer{contract: contract}, nil
}

// bindOperationValidatorSimulator binds a generic wrapper to an already deployed contract.
func bindOperationValidatorSimulator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OperationValidatorSimulatorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OperationValidatorSimulator *OperationValidatorSimulatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OperationValidatorSimulator.Contract.OperationValidatorSimulatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OperationValidatorSimulator *OperationValidatorSimulatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperationValidatorSimulator.Contract.OperationValidatorSimulatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OperationValidatorSimulator *OperationValidatorSimulatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OperationValidatorSimulator.Contract.OperationValidatorSimulatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OperationValidatorSimulator *OperationValidatorSimulatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OperationValidatorSimulator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OperationValidatorSimulator *OperationValidatorSimulatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperationValidatorSimulator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OperationValidatorSimulator *OperationValidatorSimulatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OperationValidatorSimulator.Contract.contract.Transact(opts, method, params...)
}

// SimulateOperation is a free data retrieval call binding the contract method 0x052e7f32.
//
// Solidity: function simulateOperation(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext, address _endorser, uint256 _calldataGas) view returns((bool,bool) result)
func (_OperationValidatorSimulator *OperationValidatorSimulatorCaller) SimulateOperation(opts *bind.CallOpts, _entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool, _endorser common.Address, _calldataGas *big.Int) (OperationValidatorSimulatorSimulationResult, error) {
	var out []interface{}
	err := _OperationValidatorSimulator.contract.Call(opts, &out, "simulateOperation", _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext, _endorser, _calldataGas)

	if err != nil {
		return *new(OperationValidatorSimulatorSimulationResult), err
	}

	out0 := *abi.ConvertType(out[0], new(OperationValidatorSimulatorSimulationResult)).(*OperationValidatorSimulatorSimulationResult)

	return out0, err

}

// SimulateOperation is a free data retrieval call binding the contract method 0x052e7f32.
//
// Solidity: function simulateOperation(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext, address _endorser, uint256 _calldataGas) view returns((bool,bool) result)
func (_OperationValidatorSimulator *OperationValidatorSimulatorSession) SimulateOperation(_entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool, _endorser common.Address, _calldataGas *big.Int) (OperationValidatorSimulatorSimulationResult, error) {
	return _OperationValidatorSimulator.Contract.SimulateOperation(&_OperationValidatorSimulator.CallOpts, _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext, _endorser, _calldataGas)
}

// SimulateOperation is a free data retrieval call binding the contract method 0x052e7f32.
//
// Solidity: function simulateOperation(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext, address _endorser, uint256 _calldataGas) view returns((bool,bool) result)
func (_OperationValidatorSimulator *OperationValidatorSimulatorCallerSession) SimulateOperation(_entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool, _endorser common.Address, _calldataGas *big.Int) (OperationValidatorSimulatorSimulationResult, error) {
	return _OperationValidatorSimulator.Contract.SimulateOperation(&_OperationValidatorSimulator.CallOpts, _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext, _endorser, _calldataGas)
}

// SafeExecute is a paid mutator transaction binding the contract method 0x3e505d4d.
//
// Solidity: function safeExecute(address _entrypoint, bytes _data, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, uint256 _calldataGas) returns()
func (_OperationValidatorSimulator *OperationValidatorSimulatorTransactor) SafeExecute(opts *bind.TransactOpts, _entrypoint common.Address, _data []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _calldataGas *big.Int) (*types.Transaction, error) {
	return _OperationValidatorSimulator.contract.Transact(opts, "safeExecute", _entrypoint, _data, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _calldataGas)
}

// SafeExecute is a paid mutator transaction binding the contract method 0x3e505d4d.
//
// Solidity: function safeExecute(address _entrypoint, bytes _data, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, uint256 _calldataGas) returns()
func (_OperationValidatorSimulator *OperationValidatorSimulatorSession) SafeExecute(_entrypoint common.Address, _data []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _calldataGas *big.Int) (*types.Transaction, error) {
	return _OperationValidatorSimulator.Contract.SafeExecute(&_OperationValidatorSimulator.TransactOpts, _entrypoint, _data, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _calldataGas)
}

// SafeExecute is a paid mutator transaction binding the contract method 0x3e505d4d.
//
// Solidity: function safeExecute(address _entrypoint, bytes _data, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, uint256 _calldataGas) returns()
func (_OperationValidatorSimulator *OperationValidatorSimulatorTransactorSession) SafeExecute(_entrypoint common.Address, _data []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _calldataGas *big.Int) (*types.Transaction, error) {
	return _OperationValidatorSimulator.Contract.SafeExecute(&_OperationValidatorSimulator.TransactOpts, _entrypoint, _data, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _calldataGas)
}
