// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abivalidator

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

// IEndorserOperation is an auto generated low-level Go binding around an user-defined struct.
type IEndorserOperation struct {
	Entrypoint             common.Address
	Data                   []byte
	EndorserCallData       []byte
	FixedGas               *big.Int
	GasLimit               *big.Int
	MaxFeePerGas           *big.Int
	MaxPriorityFeePerGas   *big.Int
	FeeToken               common.Address
	FeeScalingFactor       *big.Int
	FeeNormalizationFactor *big.Int
	HasUntrustedContext    bool
}

// OperationValidatorSimulationResult is an auto generated low-level Go binding around an user-defined struct.
type OperationValidatorSimulationResult struct {
	Payment *big.Int
	GasUsed *big.Int
}

// OperationValidatorMetaData contains all meta data concerning the OperationValidator contract.
var OperationValidatorMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"simulateOperation\",\"inputs\":[{\"name\":\"_op\",\"type\":\"tuple\",\"internalType\":\"structIEndorser.Operation\",\"components\":[{\"name\":\"entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"endorserCallData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"fixedGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"hasUntrustedContext\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[{\"name\":\"result\",\"type\":\"tuple\",\"internalType\":\"structOperationValidator.SimulationResult\",\"components\":[{\"name\":\"payment\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gasUsed\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"}]",
}

// OperationValidatorABI is the input ABI used to generate the binding from.
// Deprecated: Use OperationValidatorMetaData.ABI instead.
var OperationValidatorABI = OperationValidatorMetaData.ABI

// OperationValidator is an auto generated Go binding around an Ethereum contract.
type OperationValidator struct {
	OperationValidatorCaller     // Read-only binding to the contract
	OperationValidatorTransactor // Write-only binding to the contract
	OperationValidatorFilterer   // Log filterer for contract events
}

// OperationValidatorCaller is an auto generated read-only Go binding around an Ethereum contract.
type OperationValidatorCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationValidatorTransactor is an auto generated write-only Go binding around an Ethereum contract.
type OperationValidatorTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationValidatorFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OperationValidatorFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationValidatorSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OperationValidatorSession struct {
	Contract     *OperationValidator // Generic contract binding to set the session for
	CallOpts     bind.CallOpts       // Call options to use throughout this session
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// OperationValidatorCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OperationValidatorCallerSession struct {
	Contract *OperationValidatorCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts             // Call options to use throughout this session
}

// OperationValidatorTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OperationValidatorTransactorSession struct {
	Contract     *OperationValidatorTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts             // Transaction auth options to use throughout this session
}

// OperationValidatorRaw is an auto generated low-level Go binding around an Ethereum contract.
type OperationValidatorRaw struct {
	Contract *OperationValidator // Generic contract binding to access the raw methods on
}

// OperationValidatorCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OperationValidatorCallerRaw struct {
	Contract *OperationValidatorCaller // Generic read-only contract binding to access the raw methods on
}

// OperationValidatorTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OperationValidatorTransactorRaw struct {
	Contract *OperationValidatorTransactor // Generic write-only contract binding to access the raw methods on
}

// NewOperationValidator creates a new instance of OperationValidator, bound to a specific deployed contract.
func NewOperationValidator(address common.Address, backend bind.ContractBackend) (*OperationValidator, error) {
	contract, err := bindOperationValidator(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OperationValidator{OperationValidatorCaller: OperationValidatorCaller{contract: contract}, OperationValidatorTransactor: OperationValidatorTransactor{contract: contract}, OperationValidatorFilterer: OperationValidatorFilterer{contract: contract}}, nil
}

// NewOperationValidatorCaller creates a new read-only instance of OperationValidator, bound to a specific deployed contract.
func NewOperationValidatorCaller(address common.Address, caller bind.ContractCaller) (*OperationValidatorCaller, error) {
	contract, err := bindOperationValidator(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OperationValidatorCaller{contract: contract}, nil
}

// NewOperationValidatorTransactor creates a new write-only instance of OperationValidator, bound to a specific deployed contract.
func NewOperationValidatorTransactor(address common.Address, transactor bind.ContractTransactor) (*OperationValidatorTransactor, error) {
	contract, err := bindOperationValidator(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OperationValidatorTransactor{contract: contract}, nil
}

// NewOperationValidatorFilterer creates a new log filterer instance of OperationValidator, bound to a specific deployed contract.
func NewOperationValidatorFilterer(address common.Address, filterer bind.ContractFilterer) (*OperationValidatorFilterer, error) {
	contract, err := bindOperationValidator(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OperationValidatorFilterer{contract: contract}, nil
}

// bindOperationValidator binds a generic wrapper to an already deployed contract.
func bindOperationValidator(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OperationValidatorMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OperationValidator *OperationValidatorRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OperationValidator.Contract.OperationValidatorCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OperationValidator *OperationValidatorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperationValidator.Contract.OperationValidatorTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OperationValidator *OperationValidatorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OperationValidator.Contract.OperationValidatorTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OperationValidator *OperationValidatorCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OperationValidator.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OperationValidator *OperationValidatorTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperationValidator.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OperationValidator *OperationValidatorTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OperationValidator.Contract.contract.Transact(opts, method, params...)
}

// SimulateOperation is a free data retrieval call binding the contract method 0xf0bf1609.
//
// Solidity: function simulateOperation((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) _op) view returns((uint256,uint256) result)
func (_OperationValidator *OperationValidatorCaller) SimulateOperation(opts *bind.CallOpts, _op IEndorserOperation) (OperationValidatorSimulationResult, error) {
	var out []interface{}
	err := _OperationValidator.contract.Call(opts, &out, "simulateOperation", _op)

	if err != nil {
		return *new(OperationValidatorSimulationResult), err
	}

	out0 := *abi.ConvertType(out[0], new(OperationValidatorSimulationResult)).(*OperationValidatorSimulationResult)

	return out0, err

}

// SimulateOperation is a free data retrieval call binding the contract method 0xf0bf1609.
//
// Solidity: function simulateOperation((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) _op) view returns((uint256,uint256) result)
func (_OperationValidator *OperationValidatorSession) SimulateOperation(_op IEndorserOperation) (OperationValidatorSimulationResult, error) {
	return _OperationValidator.Contract.SimulateOperation(&_OperationValidator.CallOpts, _op)
}

// SimulateOperation is a free data retrieval call binding the contract method 0xf0bf1609.
//
// Solidity: function simulateOperation((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) _op) view returns((uint256,uint256) result)
func (_OperationValidator *OperationValidatorCallerSession) SimulateOperation(_op IEndorserOperation) (OperationValidatorSimulationResult, error) {
	return _OperationValidator.Contract.SimulateOperation(&_OperationValidator.CallOpts, _op)
}
