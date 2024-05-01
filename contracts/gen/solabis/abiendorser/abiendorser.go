// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package abiendorser

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

// EndorserMetaData contains all meta data concerning the Endorser contract.
var EndorserMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"isOperationReady\",\"inputs\":[{\"name\":\"operation\",\"type\":\"tuple\",\"internalType\":\"structIEndorser.Operation\",\"components\":[{\"name\":\"entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"endorserCallData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"fixedGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"hasUntrustedContext\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[{\"name\":\"readiness\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"globalDependency\",\"type\":\"tuple\",\"internalType\":\"structIEndorser.GlobalDependency\",\"components\":[{\"name\":\"baseFee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"blobBaseFee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"chainId\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"coinBase\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"difficulty\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"gasLimit\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"number\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"timestamp\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txOrigin\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txGasPrice\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"maxBlockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxBlockTimestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"dependencies\",\"type\":\"tuple[]\",\"internalType\":\"structIEndorser.Dependency[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"balance\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"code\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"nonce\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"allSlots\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"slots\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"constraints\",\"type\":\"tuple[]\",\"internalType\":\"structIEndorser.Constraint[]\",\"components\":[{\"name\":\"slot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"minValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"maxValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"simulationSettings\",\"inputs\":[{\"name\":\"operation\",\"type\":\"tuple\",\"internalType\":\"structIEndorser.Operation\",\"components\":[{\"name\":\"entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"endorserCallData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"fixedGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"feeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"feeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"hasUntrustedContext\",\"type\":\"bool\",\"internalType\":\"bool\"}]}],\"outputs\":[{\"name\":\"replacements\",\"type\":\"tuple[]\",\"internalType\":\"structIEndorser.Replacement[]\",\"components\":[{\"name\":\"oldAddr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"newAddr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"slots\",\"type\":\"tuple[]\",\"internalType\":\"structIEndorser.SlotReplacement[]\",\"components\":[{\"name\":\"slot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"value\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}],\"stateMutability\":\"view\"},{\"type\":\"event\",\"name\":\"UntrustedEnded\",\"inputs\":[],\"anonymous\":false},{\"type\":\"event\",\"name\":\"UntrustedStarted\",\"inputs\":[],\"anonymous\":false}]",
}

// EndorserABI is the input ABI used to generate the binding from.
// Deprecated: Use EndorserMetaData.ABI instead.
var EndorserABI = EndorserMetaData.ABI

// Endorser is an auto generated Go binding around an Ethereum contract.
type Endorser struct {
	EndorserCaller     // Read-only binding to the contract
	EndorserTransactor // Write-only binding to the contract
	EndorserFilterer   // Log filterer for contract events
}

// EndorserCaller is an auto generated read-only Go binding around an Ethereum contract.
type EndorserCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EndorserTransactor is an auto generated write-only Go binding around an Ethereum contract.
type EndorserTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EndorserFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type EndorserFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// EndorserSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type EndorserSession struct {
	Contract     *Endorser         // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// EndorserCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type EndorserCallerSession struct {
	Contract *EndorserCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts   // Call options to use throughout this session
}

// EndorserTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type EndorserTransactorSession struct {
	Contract     *EndorserTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts   // Transaction auth options to use throughout this session
}

// EndorserRaw is an auto generated low-level Go binding around an Ethereum contract.
type EndorserRaw struct {
	Contract *Endorser // Generic contract binding to access the raw methods on
}

// EndorserCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type EndorserCallerRaw struct {
	Contract *EndorserCaller // Generic read-only contract binding to access the raw methods on
}

// EndorserTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type EndorserTransactorRaw struct {
	Contract *EndorserTransactor // Generic write-only contract binding to access the raw methods on
}

// NewEndorser creates a new instance of Endorser, bound to a specific deployed contract.
func NewEndorser(address common.Address, backend bind.ContractBackend) (*Endorser, error) {
	contract, err := bindEndorser(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Endorser{EndorserCaller: EndorserCaller{contract: contract}, EndorserTransactor: EndorserTransactor{contract: contract}, EndorserFilterer: EndorserFilterer{contract: contract}}, nil
}

// NewEndorserCaller creates a new read-only instance of Endorser, bound to a specific deployed contract.
func NewEndorserCaller(address common.Address, caller bind.ContractCaller) (*EndorserCaller, error) {
	contract, err := bindEndorser(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &EndorserCaller{contract: contract}, nil
}

// NewEndorserTransactor creates a new write-only instance of Endorser, bound to a specific deployed contract.
func NewEndorserTransactor(address common.Address, transactor bind.ContractTransactor) (*EndorserTransactor, error) {
	contract, err := bindEndorser(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &EndorserTransactor{contract: contract}, nil
}

// NewEndorserFilterer creates a new log filterer instance of Endorser, bound to a specific deployed contract.
func NewEndorserFilterer(address common.Address, filterer bind.ContractFilterer) (*EndorserFilterer, error) {
	contract, err := bindEndorser(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &EndorserFilterer{contract: contract}, nil
}

// bindEndorser binds a generic wrapper to an already deployed contract.
func bindEndorser(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := EndorserMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, *parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Endorser *EndorserRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Endorser.Contract.EndorserCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Endorser *EndorserRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Endorser.Contract.EndorserTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Endorser *EndorserRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Endorser.Contract.EndorserTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Endorser *EndorserCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _Endorser.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Endorser *EndorserTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Endorser.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Endorser *EndorserTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Endorser.Contract.contract.Transact(opts, method, params...)
}

// IsOperationReady is a free data retrieval call binding the contract method 0x59197a0f.
//
// Solidity: function isOperationReady((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) operation) view returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_Endorser *EndorserCaller) IsOperationReady(opts *bind.CallOpts, operation IEndorserOperation) (struct {
	Readiness        bool
	GlobalDependency IEndorserGlobalDependency
	Dependencies     []IEndorserDependency
}, error) {
	var out []interface{}
	err := _Endorser.contract.Call(opts, &out, "isOperationReady", operation)

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
// Solidity: function isOperationReady((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) operation) view returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_Endorser *EndorserSession) IsOperationReady(operation IEndorserOperation) (struct {
	Readiness        bool
	GlobalDependency IEndorserGlobalDependency
	Dependencies     []IEndorserDependency
}, error) {
	return _Endorser.Contract.IsOperationReady(&_Endorser.CallOpts, operation)
}

// IsOperationReady is a free data retrieval call binding the contract method 0x59197a0f.
//
// Solidity: function isOperationReady((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) operation) view returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_Endorser *EndorserCallerSession) IsOperationReady(operation IEndorserOperation) (struct {
	Readiness        bool
	GlobalDependency IEndorserGlobalDependency
	Dependencies     []IEndorserDependency
}, error) {
	return _Endorser.Contract.IsOperationReady(&_Endorser.CallOpts, operation)
}

// SimulationSettings is a free data retrieval call binding the contract method 0x338c553c.
//
// Solidity: function simulationSettings((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) operation) view returns((address,address,(bytes32,bytes32)[])[] replacements)
func (_Endorser *EndorserCaller) SimulationSettings(opts *bind.CallOpts, operation IEndorserOperation) ([]IEndorserReplacement, error) {
	var out []interface{}
	err := _Endorser.contract.Call(opts, &out, "simulationSettings", operation)

	if err != nil {
		return *new([]IEndorserReplacement), err
	}

	out0 := *abi.ConvertType(out[0], new([]IEndorserReplacement)).(*[]IEndorserReplacement)

	return out0, err

}

// SimulationSettings is a free data retrieval call binding the contract method 0x338c553c.
//
// Solidity: function simulationSettings((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) operation) view returns((address,address,(bytes32,bytes32)[])[] replacements)
func (_Endorser *EndorserSession) SimulationSettings(operation IEndorserOperation) ([]IEndorserReplacement, error) {
	return _Endorser.Contract.SimulationSettings(&_Endorser.CallOpts, operation)
}

// SimulationSettings is a free data retrieval call binding the contract method 0x338c553c.
//
// Solidity: function simulationSettings((address,bytes,bytes,uint256,uint256,uint256,uint256,address,uint256,uint256,bool) operation) view returns((address,address,(bytes32,bytes32)[])[] replacements)
func (_Endorser *EndorserCallerSession) SimulationSettings(operation IEndorserOperation) ([]IEndorserReplacement, error) {
	return _Endorser.Contract.SimulationSettings(&_Endorser.CallOpts, operation)
}

// EndorserUntrustedEndedIterator is returned from FilterUntrustedEnded and is used to iterate over the raw logs and unpacked data for UntrustedEnded events raised by the Endorser contract.
type EndorserUntrustedEndedIterator struct {
	Event *EndorserUntrustedEnded // Event containing the contract specifics and raw log

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
func (it *EndorserUntrustedEndedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EndorserUntrustedEnded)
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
		it.Event = new(EndorserUntrustedEnded)
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
func (it *EndorserUntrustedEndedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EndorserUntrustedEndedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EndorserUntrustedEnded represents a UntrustedEnded event raised by the Endorser contract.
type EndorserUntrustedEnded struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterUntrustedEnded is a free log retrieval operation binding the contract event 0xe59e021ea70d7129da268f7f72da44741fc74a90aa5dde2db22588e4c4e3e34d.
//
// Solidity: event UntrustedEnded()
func (_Endorser *EndorserFilterer) FilterUntrustedEnded(opts *bind.FilterOpts) (*EndorserUntrustedEndedIterator, error) {

	logs, sub, err := _Endorser.contract.FilterLogs(opts, "UntrustedEnded")
	if err != nil {
		return nil, err
	}
	return &EndorserUntrustedEndedIterator{contract: _Endorser.contract, event: "UntrustedEnded", logs: logs, sub: sub}, nil
}

// WatchUntrustedEnded is a free log subscription operation binding the contract event 0xe59e021ea70d7129da268f7f72da44741fc74a90aa5dde2db22588e4c4e3e34d.
//
// Solidity: event UntrustedEnded()
func (_Endorser *EndorserFilterer) WatchUntrustedEnded(opts *bind.WatchOpts, sink chan<- *EndorserUntrustedEnded) (event.Subscription, error) {

	logs, sub, err := _Endorser.contract.WatchLogs(opts, "UntrustedEnded")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EndorserUntrustedEnded)
				if err := _Endorser.contract.UnpackLog(event, "UntrustedEnded", log); err != nil {
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
func (_Endorser *EndorserFilterer) ParseUntrustedEnded(log types.Log) (*EndorserUntrustedEnded, error) {
	event := new(EndorserUntrustedEnded)
	if err := _Endorser.contract.UnpackLog(event, "UntrustedEnded", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// EndorserUntrustedStartedIterator is returned from FilterUntrustedStarted and is used to iterate over the raw logs and unpacked data for UntrustedStarted events raised by the Endorser contract.
type EndorserUntrustedStartedIterator struct {
	Event *EndorserUntrustedStarted // Event containing the contract specifics and raw log

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
func (it *EndorserUntrustedStartedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(EndorserUntrustedStarted)
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
		it.Event = new(EndorserUntrustedStarted)
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
func (it *EndorserUntrustedStartedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *EndorserUntrustedStartedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// EndorserUntrustedStarted represents a UntrustedStarted event raised by the Endorser contract.
type EndorserUntrustedStarted struct {
	Raw types.Log // Blockchain specific contextual infos
}

// FilterUntrustedStarted is a free log retrieval operation binding the contract event 0x5e802b414aa8d12c8eb59955acbf0f9ce42afc2842ee01c46d7053c94528687a.
//
// Solidity: event UntrustedStarted()
func (_Endorser *EndorserFilterer) FilterUntrustedStarted(opts *bind.FilterOpts) (*EndorserUntrustedStartedIterator, error) {

	logs, sub, err := _Endorser.contract.FilterLogs(opts, "UntrustedStarted")
	if err != nil {
		return nil, err
	}
	return &EndorserUntrustedStartedIterator{contract: _Endorser.contract, event: "UntrustedStarted", logs: logs, sub: sub}, nil
}

// WatchUntrustedStarted is a free log subscription operation binding the contract event 0x5e802b414aa8d12c8eb59955acbf0f9ce42afc2842ee01c46d7053c94528687a.
//
// Solidity: event UntrustedStarted()
func (_Endorser *EndorserFilterer) WatchUntrustedStarted(opts *bind.WatchOpts, sink chan<- *EndorserUntrustedStarted) (event.Subscription, error) {

	logs, sub, err := _Endorser.contract.WatchLogs(opts, "UntrustedStarted")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(EndorserUntrustedStarted)
				if err := _Endorser.contract.UnpackLog(event, "UntrustedStarted", log); err != nil {
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
func (_Endorser *EndorserFilterer) ParseUntrustedStarted(log types.Log) (*EndorserUntrustedStarted, error) {
	event := new(EndorserUntrustedStarted)
	if err := _Endorser.contract.UnpackLog(event, "UntrustedStarted", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
