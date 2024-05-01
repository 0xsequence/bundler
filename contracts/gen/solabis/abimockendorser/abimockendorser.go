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

// IEndorserConstraint is an auto generated low-level Go binding around an user-defined struct.
type IEndorserConstraint struct {
	Slot     [32]byte
	MinValue [32]byte
	MaxValue [32]byte
}

// IEndorserDependency is an auto generated low-level Go binding around an user-defined struct.
type IEndorserDependency struct {
	Addr        common.Address
	Balance     bool
	Code        bool
	Nonce       bool
	AllSlots    bool
	Slots       [][32]byte
	Constraints []IEndorserConstraint
}

// IEndorserGlobalDependency is an auto generated low-level Go binding around an user-defined struct.
type IEndorserGlobalDependency struct {
	BaseFee           bool
	BlobBaseFee       bool
	ChainId           bool
	CoinBase          bool
	Difficulty        bool
	GasLimit          bool
	Number            bool
	Timestamp         bool
	TxOrigin          bool
	TxGasPrice        bool
	MaxBlockNumber    *big.Int
	MaxBlockTimestamp *big.Int
}

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

// IEndorserReplacement is an auto generated low-level Go binding around an user-defined struct.
type IEndorserReplacement struct {
	OldAddr common.Address
	NewAddr common.Address
	Slots   []IEndorserSlotReplacement
}

// IEndorserSlotReplacement is an auto generated low-level Go binding around an user-defined struct.
type IEndorserSlotReplacement struct {
	Slot  [32]byte
	Value [32]byte
}

// MockEndorserMetaData contains all meta data concerning the MockEndorser contract.
var MockEndorserMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"encodeEndorserCalldata\",\"inputs\":[{\"name\":\"_readiness\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"_globalDependency\",\"type\":\"tuple\",\"internalType\":\"structIEndorser.GlobalDependency\",\"components\":[{\"name\":\"baseFee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"blobBaseFee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"chainId\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"coinBase\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"difficulty\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"gasLimit\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"number\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"timestamp\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txOrigin\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txGasPrice\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"maxBlockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxBlockTimestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"_dependencies\",\"type\":\"tuple[]\",\"internalType\":\"structIEndorser.Dependency[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"balance\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"code\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"nonce\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"allSlots\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"slots\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"constraints\",\"type\":\"tuple[]\",\"internalType\":\"structIEndorser.Constraint[]\",\"components\":[{\"name\":\"slot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"minValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"maxValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"isOperationReady\",\"inputs\":[{\"name\":\"_op\",\"type\":\"tuple\",\"internalType\":\"structIEndorser.Operation\",\"components\":[{\"name\":\"entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"endorserCallData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"fixedGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"hasUntrustedContext\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[{\"name\":\"readiness\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"globalDependency\",\"type\":\"tuple\",\"internalType\":\"structIEndorser.GlobalDependency\",\"components\":[{\"name\":\"baseFee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"blobBaseFee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"chainId\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"coinBase\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"difficulty\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"gasLimit\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"number\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"timestamp\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txOrigin\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txGasPrice\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"maxBlockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxBlockTimestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"dependencies\",\"type\":\"tuple[]\",\"internalType\":\"structIEndorser.Dependency[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"balance\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"code\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"nonce\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"allSlots\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"slots\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"constraints\",\"type\":\"tuple[]\",\"internalType\":\"structIEndorser.Constraint[]\",\"components\":[{\"name\":\"slot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"minValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"maxValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}],\"stateMutability\":\"pure\"},{\"type\":\"function\",\"name\":\"simulationSettings\",\"inputs\":[{\"name\":\"\",\"type\":\"tuple\",\"internalType\":\"structIEndorser.Operation\",\"components\":[{\"name\":\"entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"endorserCallData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"fixedGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"hasUntrustedContext\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[{\"name\":\"replacements\",\"type\":\"tuple[]\",\"internalType\":\"structIEndorser.Replacement[]\",\"components\":[{\"name\":\"oldAddr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"newAddr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"slots\",\"type\":\"tuple[]\",\"internalType\":\"structIEndorser.SlotReplacement[]\",\"components\":[{\"name\":\"slot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"value\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}],\"stateMutability\":\"pure\"},{\"type\":\"event\",\"name\":\"UntrustedEnded\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UntrustedStarted\",\"inputs\":[],\"anonymous\":false}]",
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
func (_MockEndorser *MockEndorserCaller) EncodeEndorserCalldata(opts *bind.CallOpts, _readiness bool, _globalDependency IEndorserGlobalDependency, _dependencies []IEndorserDependency) ([]byte, error) {
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
func (_MockEndorser *MockEndorserSession) EncodeEndorserCalldata(_readiness bool, _globalDependency IEndorserGlobalDependency, _dependencies []IEndorserDependency) ([]byte, error) {
	return _MockEndorser.Contract.EncodeEndorserCalldata(&_MockEndorser.CallOpts, _readiness, _globalDependency, _dependencies)
}

// EncodeEndorserCalldata is a free data retrieval call binding the contract method 0xee91172b.
//
// Solidity: function encodeEndorserCalldata(bool _readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) _globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] _dependencies) pure returns(bytes)
func (_MockEndorser *MockEndorserCallerSession) EncodeEndorserCalldata(_readiness bool, _globalDependency IEndorserGlobalDependency, _dependencies []IEndorserDependency) ([]byte, error) {
	return _MockEndorser.Contract.EncodeEndorserCalldata(&_MockEndorser.CallOpts, _readiness, _globalDependency, _dependencies)
}

// IsOperationReady is a free data retrieval call binding the contract method 0x59197a0f.
//
// Solidity: function isOperationReady((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) _op) pure returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_MockEndorser *MockEndorserCaller) IsOperationReady(opts *bind.CallOpts, _op IEndorserOperation) (struct {
	Readiness        bool
	GlobalDependency IEndorserGlobalDependency
	Dependencies     []IEndorserDependency
}, error) {
	var out []interface{}
	err := _MockEndorser.contract.Call(opts, &out, "isOperationReady", _op)

	outstruct := new(struct {
		Readiness        bool
		GlobalDependency IEndorserGlobalDependency
		Dependencies     []IEndorserDependency
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Readiness = *abi.ConvertType(out[0], new(bool)).(*bool)
	outstruct.GlobalDependency = *abi.ConvertType(out[1], new(IEndorserGlobalDependency)).(*IEndorserGlobalDependency)
	outstruct.Dependencies = *abi.ConvertType(out[2], new([]IEndorserDependency)).(*[]IEndorserDependency)

	return *outstruct, err

}

// IsOperationReady is a free data retrieval call binding the contract method 0x59197a0f.
//
// Solidity: function isOperationReady((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) _op) pure returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_MockEndorser *MockEndorserSession) IsOperationReady(_op IEndorserOperation) (struct {
	Readiness        bool
	GlobalDependency IEndorserGlobalDependency
	Dependencies     []IEndorserDependency
}, error) {
	return _MockEndorser.Contract.IsOperationReady(&_MockEndorser.CallOpts, _op)
}

// IsOperationReady is a free data retrieval call binding the contract method 0x59197a0f.
//
// Solidity: function isOperationReady((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) _op) pure returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_MockEndorser *MockEndorserCallerSession) IsOperationReady(_op IEndorserOperation) (struct {
	Readiness        bool
	GlobalDependency IEndorserGlobalDependency
	Dependencies     []IEndorserDependency
}, error) {
	return _MockEndorser.Contract.IsOperationReady(&_MockEndorser.CallOpts, _op)
}

// SimulationSettings is a free data retrieval call binding the contract method 0x338c553c.
//
// Solidity: function simulationSettings((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) ) pure returns((address,address,(bytes32,bytes32)[])[] replacements)
func (_MockEndorser *MockEndorserCaller) SimulationSettings(opts *bind.CallOpts, arg0 IEndorserOperation) ([]IEndorserReplacement, error) {
	var out []interface{}
	err := _MockEndorser.contract.Call(opts, &out, "simulationSettings", arg0)

	if err != nil {
		return *new([]IEndorserReplacement), err
	}

	out0 := *abi.ConvertType(out[0], new([]IEndorserReplacement)).(*[]IEndorserReplacement)

	return out0, err

}

// SimulationSettings is a free data retrieval call binding the contract method 0x338c553c.
//
// Solidity: function simulationSettings((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) ) pure returns((address,address,(bytes32,bytes32)[])[] replacements)
func (_MockEndorser *MockEndorserSession) SimulationSettings(arg0 IEndorserOperation) ([]IEndorserReplacement, error) {
	return _MockEndorser.Contract.SimulationSettings(&_MockEndorser.CallOpts, arg0)
}

// SimulationSettings is a free data retrieval call binding the contract method 0x338c553c.
//
// Solidity: function simulationSettings((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) ) pure returns((address,address,(bytes32,bytes32)[])[] replacements)
func (_MockEndorser *MockEndorserCallerSession) SimulationSettings(arg0 IEndorserOperation) ([]IEndorserReplacement, error) {
	return _MockEndorser.Contract.SimulationSettings(&_MockEndorser.CallOpts, arg0)
}

// MockEndorserUntrustedEndedIterator is returned from FilterUntrustedEnded and is used to iterate over the raw logs and unpacked data for UntrustedEnded events raised by the MockEndorser contract.
type MockEndorserUntrustedEndedIterator struct {
	Event *MockEndorserUntrustedEnded // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MockEndorserUntrustedEndedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MockEndorserUntrustedEnded)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MockEndorserUntrustedEnded)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MockEndorserUntrustedEndedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MockEndorserUntrustedEndedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MockEndorserUntrustedEnded represents a UntrustedEnded event raised by the MockEndorser contract.
type MockEndorserUntrustedEnded struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterUntrustedEnded is a free log retrieval operation binding the contract event 0xe59e021ea70d7129da268f7f72da44741fc74a90aa5dde2db22588e4c4e3e34d.
//
// Solidity: event UntrustedEnded()
func (_MockEndorser *MockEndorserFilterer) FilterUntrustedEnded(opts *bind.FilterOpts) (*MockEndorserUntrustedEndedIterator, error) {

	logs, sub, err := _MockEndorser.contract.FilterLogs(opts, "UntrustedEnded")
	if err != nil {
		return nil, err
	}
	return &MockEndorserUntrustedEndedIterator{contract: _MockEndorser.contract, event: "UntrustedEnded", logs: logs, sub: sub}, nil
}

// WatchUntrustedEnded is a free log subscription operation binding the contract event 0xe59e021ea70d7129da268f7f72da44741fc74a90aa5dde2db22588e4c4e3e34d.
//
// Solidity: event UntrustedEnded()
func (_MockEndorser *MockEndorserFilterer) WatchUntrustedEnded(opts *bind.WatchOpts, sink chan<- *MockEndorserUntrustedEnded) (event.Subscription, error) {

	logs, sub, err := _MockEndorser.contract.WatchLogs(opts, "UntrustedEnded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MockEndorserUntrustedEnded)
				if err := _MockEndorser.contract.UnpackLog(event, "UntrustedEnded", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUntrustedEnded is a log parse operation binding the contract event 0xe59e021ea70d7129da268f7f72da44741fc74a90aa5dde2db22588e4c4e3e34d.
//
// Solidity: event UntrustedEnded()
func (_MockEndorser *MockEndorserFilterer) ParseUntrustedEnded(log types.Log) (*MockEndorserUntrustedEnded, error) {
	event := new(MockEndorserUntrustedEnded)
	if err := _MockEndorser.contract.UnpackLog(event, "UntrustedEnded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// MockEndorserUntrustedStartedIterator is returned from FilterUntrustedStarted and is used to iterate over the raw logs and unpacked data for UntrustedStarted events raised by the MockEndorser contract.
type MockEndorserUntrustedStartedIterator struct {
	Event *MockEndorserUntrustedStarted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *MockEndorserUntrustedStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(MockEndorserUntrustedStarted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(MockEndorserUntrustedStarted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *MockEndorserUntrustedStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *MockEndorserUntrustedStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// MockEndorserUntrustedStarted represents a UntrustedStarted event raised by the MockEndorser contract.
type MockEndorserUntrustedStarted struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterUntrustedStarted is a free log retrieval operation binding the contract event 0x5e802b414aa8d12c8eb59955acbf0f9ce42afc2842ee01c46d7053c94528687a.
//
// Solidity: event UntrustedStarted()
func (_MockEndorser *MockEndorserFilterer) FilterUntrustedStarted(opts *bind.FilterOpts) (*MockEndorserUntrustedStartedIterator, error) {

	logs, sub, err := _MockEndorser.contract.FilterLogs(opts, "UntrustedStarted")
	if err != nil {
		return nil, err
	}
	return &MockEndorserUntrustedStartedIterator{contract: _MockEndorser.contract, event: "UntrustedStarted", logs: logs, sub: sub}, nil
}

// WatchUntrustedStarted is a free log subscription operation binding the contract event 0x5e802b414aa8d12c8eb59955acbf0f9ce42afc2842ee01c46d7053c94528687a.
//
// Solidity: event UntrustedStarted()
func (_MockEndorser *MockEndorserFilterer) WatchUntrustedStarted(opts *bind.WatchOpts, sink chan<- *MockEndorserUntrustedStarted) (event.Subscription, error) {

	logs, sub, err := _MockEndorser.contract.WatchLogs(opts, "UntrustedStarted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(MockEndorserUntrustedStarted)
				if err := _MockEndorser.contract.UnpackLog(event, "UntrustedStarted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseUntrustedStarted is a log parse operation binding the contract event 0x5e802b414aa8d12c8eb59955acbf0f9ce42afc2842ee01c46d7053c94528687a.
//
// Solidity: event UntrustedStarted()
func (_MockEndorser *MockEndorserFilterer) ParseUntrustedStarted(log types.Log) (*MockEndorserUntrustedStarted, error) {
	event := new(MockEndorserUntrustedStarted)
	if err := _MockEndorser.contract.UnpackLog(event, "UntrustedStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
