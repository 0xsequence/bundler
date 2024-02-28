// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abimockendorser

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

// EndorserConstraint is an auto generated low-level Go binding around an user-defined struct.
type EndorserConstraint struct {
	Slot     [32]byte
	MinValue [32]byte
	MaxValue [32]byte
}

// EndorserDependency is an auto generated low-level Go binding around an user-defined struct.
type EndorserDependency struct {
	Addr        common.Address
	Balance     bool
	Code        bool
	Nonce       bool
	AllSlots    bool
	Slots       [][32]byte
	Constraints []EndorserConstraint
}

// EndorserGlobalDependency is an auto generated low-level Go binding around an user-defined struct.
type EndorserGlobalDependency struct {
	Basefee           bool
	Blobbasefee       bool
	Chainid           bool
	Coinbase          bool
	Difficulty        bool
	GasLimit          bool
	Number            bool
	Timestamp         bool
	TxOrigin          bool
	TxGasPrice        bool
	MaxBlockNumber    *big.Int
	MaxBlockTimestamp *big.Int
}

// MockEndorserMetaData contains all meta data concerning the MockEndorser contract.
var MockEndorserMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"encodeEndorserCalldata\",\"inputs\":[{\"name\":\"_readiness\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"_globalDependency\",\"type\":\"tuple\",\"internalType\":\"structEndorser.GlobalDependency\",\"components\":[{\"name\":\"basefee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"blobbasefee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"chainid\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"coinbase\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"difficulty\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"gasLimit\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"number\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"timestamp\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txOrigin\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txGasPrice\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"maxBlockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxBlockTimestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"_dependencies\",\"type\":\"tuple[]\",\"internalType\":\"structEndorser.Dependency[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"balance\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"code\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"nonce\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"allSlots\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"slots\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"constraints\",\"type\":\"tuple[]\",\"internalType\":\"structEndorser.Constraint[]\",\"components\":[{\"name\":\"slot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"minValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"maxValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"isOperationReady\",\"inputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_endorserCallData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[{\"name\":\"readiness\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"globalDependency\",\"type\":\"tuple\",\"internalType\":\"structEndorser.GlobalDependency\",\"components\":[{\"name\":\"basefee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"blobbasefee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"chainid\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"coinbase\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"difficulty\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"gasLimit\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"number\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"timestamp\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txOrigin\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txGasPrice\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"maxBlockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxBlockTimestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"dependencies\",\"type\":\"tuple[]\",\"internalType\":\"structEndorser.Dependency[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"balance\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"code\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"nonce\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"allSlots\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"slots\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"constraints\",\"type\":\"tuple[]\",\"internalType\":\"structEndorser.Constraint[]\",\"components\":[{\"name\":\"slot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"minValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"maxValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}],\"stateMutability\":\"pure\"}]",
}

// MockEndorserABI is the input ABI used to generate the binding from.
// Deprecated: Use MockEndorserMetaData.ABI instead.
var MockEndorserABI = MockEndorserMetaData.ABI

// MockEndorser is an auto generated Go binding around an Ethereum contract.
type MockEndorser struct {
	MockEndorserCaller     // Read-only binding to the contract
	MockEndorserTransactor // Write-only binding to the contract
	MockEndorserFilterer   // Log filterer for contract events
}

// MockEndorserCaller is an auto generated read-only Go binding around an Ethereum contract.
type MockEndorserCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockEndorserTransactor is an auto generated write-only Go binding around an Ethereum contract.
type MockEndorserTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockEndorserFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type MockEndorserFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// MockEndorserSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type MockEndorserSession struct {
	Contract     *MockEndorser     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// MockEndorserCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type MockEndorserCallerSession struct {
	Contract *MockEndorserCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// MockEndorserTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type MockEndorserTransactorSession struct {
	Contract     *MockEndorserTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// MockEndorserRaw is an auto generated low-level Go binding around an Ethereum contract.
type MockEndorserRaw struct {
	Contract *MockEndorser // Generic contract binding to access the raw methods on
}

// MockEndorserCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type MockEndorserCallerRaw struct {
	Contract *MockEndorserCaller // Generic read-only contract binding to access the raw methods on
}

// MockEndorserTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type MockEndorserTransactorRaw struct {
	Contract *MockEndorserTransactor // Generic write-only contract binding to access the raw methods on
}

// NewMockEndorser creates a new instance of MockEndorser, bound to a specific deployed contract.
func NewMockEndorser(address common.Address, backend bind.ContractBackend) (*MockEndorser, error) {
	contract, err := bindMockEndorser(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &MockEndorser{MockEndorserCaller: MockEndorserCaller{contract: contract}, MockEndorserTransactor: MockEndorserTransactor{contract: contract}, MockEndorserFilterer: MockEndorserFilterer{contract: contract}}, nil
}

// NewMockEndorserCaller creates a new read-only instance of MockEndorser, bound to a specific deployed contract.
func NewMockEndorserCaller(address common.Address, caller bind.ContractCaller) (*MockEndorserCaller, error) {
	contract, err := bindMockEndorser(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &MockEndorserCaller{contract: contract}, nil
}

// NewMockEndorserTransactor creates a new write-only instance of MockEndorser, bound to a specific deployed contract.
func NewMockEndorserTransactor(address common.Address, transactor bind.ContractTransactor) (*MockEndorserTransactor, error) {
	contract, err := bindMockEndorser(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &MockEndorserTransactor{contract: contract}, nil
}

// NewMockEndorserFilterer creates a new log filterer instance of MockEndorser, bound to a specific deployed contract.
func NewMockEndorserFilterer(address common.Address, filterer bind.ContractFilterer) (*MockEndorserFilterer, error) {
	contract, err := bindMockEndorser(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &MockEndorserFilterer{contract: contract}, nil
}

// bindMockEndorser binds a generic wrapper to an already deployed contract.
func bindMockEndorser(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := MockEndorserMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MockEndorser *MockEndorserRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MockEndorser.Contract.MockEndorserCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MockEndorser *MockEndorserRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MockEndorser.Contract.MockEndorserTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MockEndorser *MockEndorserRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MockEndorser.Contract.MockEndorserTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_MockEndorser *MockEndorserCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _MockEndorser.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_MockEndorser *MockEndorserTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _MockEndorser.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_MockEndorser *MockEndorserTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _MockEndorser.Contract.contract.Transact(opts, method, params...)
}

// EncodeEndorserCalldata is a free data retrieval call binding the contract method 0xee91172b.
//
// Solidity: function encodeEndorserCalldata(bool _readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) _globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] _dependencies) pure returns(bytes)
func (_MockEndorser *MockEndorserCaller) EncodeEndorserCalldata(opts *bind.CallOpts, _readiness bool, _globalDependency EndorserGlobalDependency, _dependencies []EndorserDependency) ([]byte, error) {
	var out []interface{}
	err := _MockEndorser.contract.Call(opts, &out, "encodeEndorserCalldata", _readiness, _globalDependency, _dependencies)

	if err != nil {
		return *new([]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([]byte)).(*[]byte)

	return out0, err

}

// EncodeEndorserCalldata is a free data retrieval call binding the contract method 0xee91172b.
//
// Solidity: function encodeEndorserCalldata(bool _readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) _globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] _dependencies) pure returns(bytes)
func (_MockEndorser *MockEndorserSession) EncodeEndorserCalldata(_readiness bool, _globalDependency EndorserGlobalDependency, _dependencies []EndorserDependency) ([]byte, error) {
	return _MockEndorser.Contract.EncodeEndorserCalldata(&_MockEndorser.CallOpts, _readiness, _globalDependency, _dependencies)
}

// EncodeEndorserCalldata is a free data retrieval call binding the contract method 0xee91172b.
//
// Solidity: function encodeEndorserCalldata(bool _readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) _globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] _dependencies) pure returns(bytes)
func (_MockEndorser *MockEndorserCallerSession) EncodeEndorserCalldata(_readiness bool, _globalDependency EndorserGlobalDependency, _dependencies []EndorserDependency) ([]byte, error) {
	return _MockEndorser.Contract.EncodeEndorserCalldata(&_MockEndorser.CallOpts, _readiness, _globalDependency, _dependencies)
}

// IsOperationReady is a free data retrieval call binding the contract method 0xc573aa04.
//
// Solidity: function isOperationReady(address , bytes , bytes _endorserCallData, uint256 , uint256 , uint256 , address , uint256 , uint256 , bool ) pure returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_MockEndorser *MockEndorserCaller) IsOperationReady(opts *bind.CallOpts, arg0 common.Address, arg1 []byte, _endorserCallData []byte, arg3 *big.Int, arg4 *big.Int, arg5 *big.Int, arg6 common.Address, arg7 *big.Int, arg8 *big.Int, arg9 bool) (struct {
	Readiness        bool
	GlobalDependency EndorserGlobalDependency
	Dependencies     []EndorserDependency
}, error) {
	var out []interface{}
	err := _MockEndorser.contract.Call(opts, &out, "isOperationReady", arg0, arg1, _endorserCallData, arg3, arg4, arg5, arg6, arg7, arg8, arg9)

	outstruct := new(struct {
		Readiness        bool
		GlobalDependency EndorserGlobalDependency
		Dependencies     []EndorserDependency
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Readiness = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.GlobalDependency = *abi.ConvertType(out[1], new(EndorserGlobalDependency)).(*EndorserGlobalDependency)
	outstruct.Dependencies = *abi.ConvertType(out[2], new([]EndorserDependency)).(*[]EndorserDependency)

	return *outstruct, err

}

// IsOperationReady is a free data retrieval call binding the contract method 0xc573aa04.
//
// Solidity: function isOperationReady(address , bytes , bytes _endorserCallData, uint256 , uint256 , uint256 , address , uint256 , uint256 , bool ) pure returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_MockEndorser *MockEndorserSession) IsOperationReady(arg0 common.Address, arg1 []byte, _endorserCallData []byte, arg3 *big.Int, arg4 *big.Int, arg5 *big.Int, arg6 common.Address, arg7 *big.Int, arg8 *big.Int, arg9 bool) (struct {
	Readiness        bool
	GlobalDependency EndorserGlobalDependency
	Dependencies     []EndorserDependency
}, error) {
	return _MockEndorser.Contract.IsOperationReady(&_MockEndorser.CallOpts, arg0, arg1, _endorserCallData, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
}

// IsOperationReady is a free data retrieval call binding the contract method 0xc573aa04.
//
// Solidity: function isOperationReady(address , bytes , bytes _endorserCallData, uint256 , uint256 , uint256 , address , uint256 , uint256 , bool ) pure returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_MockEndorser *MockEndorserCallerSession) IsOperationReady(arg0 common.Address, arg1 []byte, _endorserCallData []byte, arg3 *big.Int, arg4 *big.Int, arg5 *big.Int, arg6 common.Address, arg7 *big.Int, arg8 *big.Int, arg9 bool) (struct {
	Readiness        bool
	GlobalDependency EndorserGlobalDependency
	Dependencies     []EndorserDependency
}, error) {
	return _MockEndorser.Contract.IsOperationReady(&_MockEndorser.CallOpts, arg0, arg1, _endorserCallData, arg3, arg4, arg5, arg6, arg7, arg8, arg9)
}
