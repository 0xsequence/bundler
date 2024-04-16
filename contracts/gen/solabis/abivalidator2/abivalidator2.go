// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abivalidator2

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

// OperationValidator2SimulationResult2 is an auto generated low-level Go binding around an user-defined struct.
type OperationValidator2SimulationResult2 struct {
	Payment *big.Int
	GasUsed *big.Int
}

// OperationValidator2MetaData contains all meta data concerning the OperationValidator2 contract.
var OperationValidator2MetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"simulateOperation\",\"inputs\":[{\"name\":\"_op\",\"type\":\"tuple\",\"internalType\":\"structIEndorser.Operation\",\"components\":[{\"name\":\"entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"endorserCallData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"fixedGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"hasUntrustedContext\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[{\"name\":\"result\",\"type\":\"tuple\",\"internalType\":\"structOperationValidator2.SimulationResult2\",\"components\":[{\"name\":\"payment\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gasUsed\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]}],\"stateMutability\":\"view\"}]",
}

// OperationValidator2ABI is the input ABI used to generate the binding from.
// Deprecated: Use OperationValidator2MetaData.ABI instead.
var OperationValidator2ABI = OperationValidator2MetaData.ABI

// OperationValidator2 is an auto generated Go binding around an Ethereum contract.
type OperationValidator2 struct {
	OperationValidator2Caller     // Read-only binding to the contract
	OperationValidator2Transactor // Write-only binding to the contract
	OperationValidator2Filterer   // Log filterer for contract events
}

// OperationValidator2Caller is an auto generated read-only Go binding around an Ethereum contract.
type OperationValidator2Caller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationValidator2Transactor is an auto generated write-only Go binding around an Ethereum contract.
type OperationValidator2Transactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationValidator2Filterer is an auto generated log filtering Go binding around an Ethereum contract events.
type OperationValidator2Filterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// OperationValidator2Session is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type OperationValidator2Session struct {
	Contract     *OperationValidator2 // Generic contract binding to set the session for
	CallOpts     bind.CallOpts        // Call options to use throughout this session
	TransactOpts bind.TransactOpts    // Transaction auth options to use throughout this session
}

// OperationValidator2CallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type OperationValidator2CallerSession struct {
	Contract *OperationValidator2Caller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts              // Call options to use throughout this session
}

// OperationValidator2TransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type OperationValidator2TransactorSession struct {
	Contract     *OperationValidator2Transactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts              // Transaction auth options to use throughout this session
}

// OperationValidator2Raw is an auto generated low-level Go binding around an Ethereum contract.
type OperationValidator2Raw struct {
	Contract *OperationValidator2 // Generic contract binding to access the raw methods on
}

// OperationValidator2CallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type OperationValidator2CallerRaw struct {
	Contract *OperationValidator2Caller // Generic read-only contract binding to access the raw methods on
}

// OperationValidator2TransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type OperationValidator2TransactorRaw struct {
	Contract *OperationValidator2Transactor // Generic write-only contract binding to access the raw methods on
}

// NewOperationValidator2 creates a new instance of OperationValidator2, bound to a specific deployed contract.
func NewOperationValidator2(address common.Address, backend bind.ContractBackend) (*OperationValidator2, error) {
	contract, err := bindOperationValidator2(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &OperationValidator2{OperationValidator2Caller: OperationValidator2Caller{contract: contract}, OperationValidator2Transactor: OperationValidator2Transactor{contract: contract}, OperationValidator2Filterer: OperationValidator2Filterer{contract: contract}}, nil
}

// NewOperationValidator2Caller creates a new read-only instance of OperationValidator2, bound to a specific deployed contract.
func NewOperationValidator2Caller(address common.Address, caller bind.ContractCaller) (*OperationValidator2Caller, error) {
	contract, err := bindOperationValidator2(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &OperationValidator2Caller{contract: contract}, nil
}

// NewOperationValidator2Transactor creates a new write-only instance of OperationValidator2, bound to a specific deployed contract.
func NewOperationValidator2Transactor(address common.Address, transactor bind.ContractTransactor) (*OperationValidator2Transactor, error) {
	contract, err := bindOperationValidator2(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &OperationValidator2Transactor{contract: contract}, nil
}

// NewOperationValidator2Filterer creates a new log filterer instance of OperationValidator2, bound to a specific deployed contract.
func NewOperationValidator2Filterer(address common.Address, filterer bind.ContractFilterer) (*OperationValidator2Filterer, error) {
	contract, err := bindOperationValidator2(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &OperationValidator2Filterer{contract: contract}, nil
}

// bindOperationValidator2 binds a generic wrapper to an already deployed contract.
func bindOperationValidator2(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := OperationValidator2MetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OperationValidator2 *OperationValidator2Raw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OperationValidator2.Contract.OperationValidator2Caller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OperationValidator2 *OperationValidator2Raw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperationValidator2.Contract.OperationValidator2Transactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OperationValidator2 *OperationValidator2Raw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OperationValidator2.Contract.OperationValidator2Transactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_OperationValidator2 *OperationValidator2CallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _OperationValidator2.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_OperationValidator2 *OperationValidator2TransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _OperationValidator2.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_OperationValidator2 *OperationValidator2TransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _OperationValidator2.Contract.contract.Transact(opts, method, params...)
}

// SimulateOperation is a free data retrieval call binding the contract method 0xf0bf1609.
//
// Solidity: function simulateOperation((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) _op) view returns((uint256,uint256) result)
func (_OperationValidator2 *OperationValidator2Caller) SimulateOperation(opts *bind.CallOpts, _op IEndorserOperation) (OperationValidator2SimulationResult2, error) {
	var out []interface{}
	err := _OperationValidator2.contract.Call(opts, &out, "simulateOperation", _op)

	if err != nil {
		return *new(OperationValidator2SimulationResult2), err
	}

	out0 := *abi.ConvertType(out[0], new(OperationValidator2SimulationResult2)).(*OperationValidator2SimulationResult2)

	return out0, err

}

// SimulateOperation is a free data retrieval call binding the contract method 0xf0bf1609.
//
// Solidity: function simulateOperation((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) _op) view returns((uint256,uint256) result)
func (_OperationValidator2 *OperationValidator2Session) SimulateOperation(_op IEndorserOperation) (OperationValidator2SimulationResult2, error) {
	return _OperationValidator2.Contract.SimulateOperation(&_OperationValidator2.CallOpts, _op)
}

// SimulateOperation is a free data retrieval call binding the contract method 0xf0bf1609.
//
// Solidity: function simulateOperation((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) _op) view returns((uint256,uint256) result)
func (_OperationValidator2 *OperationValidator2CallerSession) SimulateOperation(_op IEndorserOperation) (OperationValidator2SimulationResult2, error) {
	return _OperationValidator2.Contract.SimulateOperation(&_OperationValidator2.CallOpts, _op)
}
