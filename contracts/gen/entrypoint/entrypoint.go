// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package entrypoint

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

// BundlerEntrypointSimulationResult is an auto generated low-level Go binding around an user-defined struct.
type BundlerEntrypointSimulationResult struct {
	Paid bool
	Lied bool
}

// EntrypointMetaData contains all meta data concerning the Entrypoint contract.
var EntrypointMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"safeExecute\",\"inputs\":[{\"name\":\"_entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_baseFeeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_baseFeeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_calldataGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"simulateOperation\",\"inputs\":[{\"name\":\"_entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_endorserCallData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_baseFeeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_baseFeeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_hasUntrustedContext\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"_endorser\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_calldataGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"result\",\"type\":\"tuple\",\"internalType\":\"structBundlerEntrypoint.SimulationResult\",\"components\":[{\"name\":\"paid\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"lied\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"stateMutability\":\"nonpayable\"},{\"type\":\"error\",\"name\":\"BundlerExecutionFailed\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"BundlerUnderpaid\",\"inputs\":[{\"name\":\"_paid\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_expected\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"DIVISION_BY_ZERO\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"UNDER_OVERFLOW\",\"inputs\":[]}]",
}

// EntrypointABI is the input ABI used to generate the binding from.
// Deprecated: Use EntrypointMetaData.ABI instead.
var EntrypointABI = EntrypointMetaData.ABI

// Entrypoint is an auto generated Go binding around an Ethereum contract.
type Entrypoint struct {
	EntrypointCaller     // Read-only binding to the contract
	EntrypointTransactor // Write-only binding to the contract
	EntrypointFilterer   // Log filterer for contract events
}

// EntrypointCaller is an auto generated read-only Go binding around an Ethereum contract.
type EntrypointCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EntrypointTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EntrypointTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EntrypointFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EntrypointFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EntrypointSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EntrypointSession struct {
	Contract     *Entrypoint       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EntrypointCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EntrypointCallerSession struct {
	Contract *EntrypointCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// EntrypointTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EntrypointTransactorSession struct {
	Contract     *EntrypointTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// EntrypointRaw is an auto generated low-level Go binding around an Ethereum contract.
type EntrypointRaw struct {
	Contract *Entrypoint // Generic contract binding to access the raw methods on
}

// EntrypointCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EntrypointCallerRaw struct {
	Contract *EntrypointCaller // Generic read-only contract binding to access the raw methods on
}

// EntrypointTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EntrypointTransactorRaw struct {
	Contract *EntrypointTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEntrypoint creates a new instance of Entrypoint, bound to a specific deployed contract.
func NewEntrypoint(address common.Address, backend bind.ContractBackend) (*Entrypoint, error) {
	contract, err := bindEntrypoint(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Entrypoint{EntrypointCaller: EntrypointCaller{contract: contract}, EntrypointTransactor: EntrypointTransactor{contract: contract}, EntrypointFilterer: EntrypointFilterer{contract: contract}}, nil
}

// NewEntrypointCaller creates a new read-only instance of Entrypoint, bound to a specific deployed contract.
func NewEntrypointCaller(address common.Address, caller bind.ContractCaller) (*EntrypointCaller, error) {
	contract, err := bindEntrypoint(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EntrypointCaller{contract: contract}, nil
}

// NewEntrypointTransactor creates a new write-only instance of Entrypoint, bound to a specific deployed contract.
func NewEntrypointTransactor(address common.Address, transactor bind.ContractTransactor) (*EntrypointTransactor, error) {
	contract, err := bindEntrypoint(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EntrypointTransactor{contract: contract}, nil
}

// NewEntrypointFilterer creates a new log filterer instance of Entrypoint, bound to a specific deployed contract.
func NewEntrypointFilterer(address common.Address, filterer bind.ContractFilterer) (*EntrypointFilterer, error) {
	contract, err := bindEntrypoint(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EntrypointFilterer{contract: contract}, nil
}

// bindEntrypoint binds a generic wrapper to an already deployed contract.
func bindEntrypoint(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := EntrypointMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Entrypoint *EntrypointRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Entrypoint.Contract.EntrypointCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Entrypoint *EntrypointRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Entrypoint.Contract.EntrypointTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Entrypoint *EntrypointRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Entrypoint.Contract.EntrypointTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Entrypoint *EntrypointCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Entrypoint.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Entrypoint *EntrypointTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Entrypoint.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Entrypoint *EntrypointTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Entrypoint.Contract.contract.Transact(opts, method, params...)
}

// SafeExecute is a paid mutator transaction binding the contract method 0x3e505d4d.
//
// Solidity: function safeExecute(address _entrypoint, bytes _data, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, uint256 _calldataGas) returns()
func (_Entrypoint *EntrypointTransactor) SafeExecute(opts *bind.TransactOpts, _entrypoint common.Address, _data []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Entrypoint.contract.Transact(opts, "safeExecute", _entrypoint, _data, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _calldataGas)
}

// SafeExecute is a paid mutator transaction binding the contract method 0x3e505d4d.
//
// Solidity: function safeExecute(address _entrypoint, bytes _data, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, uint256 _calldataGas) returns()
func (_Entrypoint *EntrypointSession) SafeExecute(_entrypoint common.Address, _data []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Entrypoint.Contract.SafeExecute(&_Entrypoint.TransactOpts, _entrypoint, _data, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _calldataGas)
}

// SafeExecute is a paid mutator transaction binding the contract method 0x3e505d4d.
//
// Solidity: function safeExecute(address _entrypoint, bytes _data, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, uint256 _calldataGas) returns()
func (_Entrypoint *EntrypointTransactorSession) SafeExecute(_entrypoint common.Address, _data []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Entrypoint.Contract.SafeExecute(&_Entrypoint.TransactOpts, _entrypoint, _data, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _calldataGas)
}

// SimulateOperation is a paid mutator transaction binding the contract method 0x052e7f32.
//
// Solidity: function simulateOperation(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext, address _endorser, uint256 _calldataGas) returns((bool,bool) result)
func (_Entrypoint *EntrypointTransactor) SimulateOperation(opts *bind.TransactOpts, _entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool, _endorser common.Address, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Entrypoint.contract.Transact(opts, "simulateOperation", _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext, _endorser, _calldataGas)
}

// SimulateOperation is a paid mutator transaction binding the contract method 0x052e7f32.
//
// Solidity: function simulateOperation(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext, address _endorser, uint256 _calldataGas) returns((bool,bool) result)
func (_Entrypoint *EntrypointSession) SimulateOperation(_entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool, _endorser common.Address, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Entrypoint.Contract.SimulateOperation(&_Entrypoint.TransactOpts, _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext, _endorser, _calldataGas)
}

// SimulateOperation is a paid mutator transaction binding the contract method 0x052e7f32.
//
// Solidity: function simulateOperation(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext, address _endorser, uint256 _calldataGas) returns((bool,bool) result)
func (_Entrypoint *EntrypointTransactorSession) SimulateOperation(_entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool, _endorser common.Address, _calldataGas *big.Int) (*types.Transaction, error) {
	return _Entrypoint.Contract.SimulateOperation(&_Entrypoint.TransactOpts, _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext, _endorser, _calldataGas)
}
