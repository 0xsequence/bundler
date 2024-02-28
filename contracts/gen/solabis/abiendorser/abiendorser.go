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

// EndorserMetaData contains all meta data concerning the Endorser contract.
var EndorserMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"function\",\"name\":\"isOperationReady\",\"inputs\":[{\"name\":\"_entrypoint\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_data\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_endorserCallData\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"_gasLimit\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_maxPriorityFeePerGas\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_feeToken\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"_baseFeeScalingFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_baseFeeNormalizationFactor\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"_hasUntrustedContext\",\"type\":\"bool\",\"internalType\":\"bool\"}],\"outputs\":[{\"name\":\"readiness\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"globalDependency\",\"type\":\"tuple\",\"internalType\":\"structEndorser.GlobalDependency\",\"components\":[{\"name\":\"basefee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"blobbasefee\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"chainid\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"coinbase\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"difficulty\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"gasLimit\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"number\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"timestamp\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txOrigin\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"txGasPrice\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"maxBlockNumber\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"maxBlockTimestamp\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"name\":\"dependencies\",\"type\":\"tuple[]\",\"internalType\":\"structEndorser.Dependency[]\",\"components\":[{\"name\":\"addr\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"balance\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"code\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"nonce\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"allSlots\",\"type\":\"bool\",\"internalType\":\"bool\"},{\"name\":\"slots\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"},{\"name\":\"constraints\",\"type\":\"tuple[]\",\"internalType\":\"structEndorser.Constraint[]\",\"components\":[{\"name\":\"slot\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"minValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"maxValue\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]}]}],\"stateMutability\":\"view\"}]",
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

// IsOperationReady is a free data retrieval call binding the contract method 0xc573aa04.
//
// Solidity: function isOperationReady(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext) view returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_Endorser *EndorserCaller) IsOperationReady(opts *bind.CallOpts, _entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool) (struct {
	Readiness        bool
	GlobalDependency EndorserGlobalDependency
	Dependencies     []EndorserDependency
}, error) {
	var out []interface{}
	err := _Endorser.contract.Call(opts, &out, "isOperationReady", _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext)

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
// Solidity: function isOperationReady(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext) view returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_Endorser *EndorserSession) IsOperationReady(_entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool) (struct {
	Readiness        bool
	GlobalDependency EndorserGlobalDependency
	Dependencies     []EndorserDependency
}, error) {
	return _Endorser.Contract.IsOperationReady(&_Endorser.CallOpts, _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext)
}

// IsOperationReady is a free data retrieval call binding the contract method 0xc573aa04.
//
// Solidity: function isOperationReady(address _entrypoint, bytes _data, bytes _endorserCallData, uint256 _gasLimit, uint256 _maxFeePerGas, uint256 _maxPriorityFeePerGas, address _feeToken, uint256 _baseFeeScalingFactor, uint256 _baseFeeNormalizationFactor, bool _hasUntrustedContext) view returns(bool readiness, (bool,bool,bool,bool,bool,bool,bool,bool,bool,bool,uint256,uint256) globalDependency, (address,bool,bool,bool,bool,bytes32[],(bytes32,bytes32,bytes32)[])[] dependencies)
func (_Endorser *EndorserCallerSession) IsOperationReady(_entrypoint common.Address, _data []byte, _endorserCallData []byte, _gasLimit *big.Int, _maxFeePerGas *big.Int, _maxPriorityFeePerGas *big.Int, _feeToken common.Address, _baseFeeScalingFactor *big.Int, _baseFeeNormalizationFactor *big.Int, _hasUntrustedContext bool) (struct {
	Readiness        bool
	GlobalDependency EndorserGlobalDependency
	Dependencies     []EndorserDependency
}, error) {
	return _Endorser.Contract.IsOperationReady(&_Endorser.CallOpts, _entrypoint, _data, _endorserCallData, _gasLimit, _maxFeePerGas, _maxPriorityFeePerGas, _feeToken, _baseFeeScalingFactor, _baseFeeNormalizationFactor, _hasUntrustedContext)
}
