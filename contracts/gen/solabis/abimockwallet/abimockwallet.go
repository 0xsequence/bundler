// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abimockwallet

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

// MockWalletMetaData contains all meta data concerning the MockWallet contract.
var MockWalletMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"receive\",\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"execute\",\"inputs\":[{\"name\":\"_to\",\"type\":\"address[]\",\"internalType\":\"addresspayable[]\"},{\"name\":\"_value\",\"type\":\"uint256[]\",\"internalType\":\"uint256[]\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"}]",
}

// MockWalletABI is the input ABI used to generate the binding from.
// Deprecated: Use MockWalletMetaData.ABI instead.
var MockWalletABI = MockWalletMetaData.ABI

// MockWallet is an auto generated Go binding around an Ethereum contract.
type MockWallet struct {
	MockWalletCaller     // Read-only binding to the contract
	MockWalletTransactor // Write-only binding to the contract
	MockWalletFilterer   // Log filterer for contract events
}

// MockWalletCaller is an auto generated read-only Go binding around an Ethereum contract.
type MockWalletCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockWalletTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MockWalletTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockWalletFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MockWalletFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockWalletSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MockWalletSession struct {
	Contract     *MockWallet       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MockWalletCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MockWalletCallerSession struct {
	Contract *MockWalletCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// MockWalletTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MockWalletTransactorSession struct {
	Contract     *MockWalletTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// MockWalletRaw is an auto generated low-level Go binding around an Ethereum contract.
type MockWalletRaw struct {
	Contract *MockWallet // Generic contract binding to access the raw methods on
}

// MockWalletCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MockWalletCallerRaw struct {
	Contract *MockWalletCaller // Generic read-only contract binding to access the raw methods on
}

// MockWalletTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MockWalletTransactorRaw struct {
	Contract *MockWalletTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMockWallet creates a new instance of MockWallet, bound to a specific deployed contract.
func NewMockWallet(address common.Address, backend bind.ContractBackend) (*MockWallet, error) {
	contract, err := bindMockWallet(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MockWallet{MockWalletCaller: MockWalletCaller{contract: contract}, MockWalletTransactor: MockWalletTransactor{contract: contract}, MockWalletFilterer: MockWalletFilterer{contract: contract}}, nil
}

// NewMockWalletCaller creates a new read-only instance of MockWallet, bound to a specific deployed contract.
func NewMockWalletCaller(address common.Address, caller bind.ContractCaller) (*MockWalletCaller, error) {
	contract, err := bindMockWallet(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MockWalletCaller{contract: contract}, nil
}

// NewMockWalletTransactor creates a new write-only instance of MockWallet, bound to a specific deployed contract.
func NewMockWalletTransactor(address common.Address, transactor bind.ContractTransactor) (*MockWalletTransactor, error) {
	contract, err := bindMockWallet(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MockWalletTransactor{contract: contract}, nil
}

// NewMockWalletFilterer creates a new log filterer instance of MockWallet, bound to a specific deployed contract.
func NewMockWalletFilterer(address common.Address, filterer bind.ContractFilterer) (*MockWalletFilterer, error) {
	contract, err := bindMockWallet(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MockWalletFilterer{contract: contract}, nil
}

// bindMockWallet binds a generic wrapper to an already deployed contract.
func bindMockWallet(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MockWalletMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MockWallet *MockWalletRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MockWallet.Contract.MockWalletCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MockWallet *MockWalletRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MockWallet.Contract.MockWalletTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MockWallet *MockWalletRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MockWallet.Contract.MockWalletTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MockWallet *MockWalletCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MockWallet.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MockWallet *MockWalletTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MockWallet.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MockWallet *MockWalletTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MockWallet.Contract.contract.Transact(opts, method, params...)
}

// Execute is a paid mutator transaction binding the contract method 0x553500b4.
//
// Solidity: function execute(address[] _to, uint256[] _value) returns()
func (_MockWallet *MockWalletTransactor) Execute(opts *bind.TransactOpts, _to []common.Address, _value []*big.Int) (*types.Transaction, error) {
	return _MockWallet.contract.Transact(opts, "execute", _to, _value)
}

// Execute is a paid mutator transaction binding the contract method 0x553500b4.
//
// Solidity: function execute(address[] _to, uint256[] _value) returns()
func (_MockWallet *MockWalletSession) Execute(_to []common.Address, _value []*big.Int) (*types.Transaction, error) {
	return _MockWallet.Contract.Execute(&_MockWallet.TransactOpts, _to, _value)
}

// Execute is a paid mutator transaction binding the contract method 0x553500b4.
//
// Solidity: function execute(address[] _to, uint256[] _value) returns()
func (_MockWallet *MockWalletTransactorSession) Execute(_to []common.Address, _value []*big.Int) (*types.Transaction, error) {
	return _MockWallet.Contract.Execute(&_MockWallet.TransactOpts, _to, _value)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MockWallet *MockWalletTransactor) Receive(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MockWallet.contract.RawTransact(opts, nil) // calldata is disallowed for receive function
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MockWallet *MockWalletSession) Receive() (*types.Transaction, error) {
	return _MockWallet.Contract.Receive(&_MockWallet.TransactOpts)
}

// Receive is a paid mutator transaction binding the contract receive function.
//
// Solidity: receive() payable returns()
func (_MockWallet *MockWalletTransactorSession) Receive() (*types.Transaction, error) {
	return _MockWallet.Contract.Receive(&_MockWallet.TransactOpts)
}
