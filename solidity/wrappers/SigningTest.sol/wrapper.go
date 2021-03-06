// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package wrappers

import (
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
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// SigningTestABI is the input ABI used to generate the binding from.
const SigningTestABI = "[{\"inputs\":[{\"internalType\":\"address\",\"name\":\"_signer\",\"type\":\"address\"},{\"internalType\":\"bytes32\",\"name\":\"_theHash\",\"type\":\"bytes32\"},{\"internalType\":\"uint8\",\"name\":\"_v\",\"type\":\"uint8\"},{\"internalType\":\"bytes32\",\"name\":\"_r\",\"type\":\"bytes32\"},{\"internalType\":\"bytes32\",\"name\":\"_s\",\"type\":\"bytes32\"}],\"name\":\"checkSignature\",\"outputs\":[],\"stateMutability\":\"view\",\"type\":\"function\"}]"

// SigningTestFuncSigs maps the 4-byte function signature to its string representation.
var SigningTestFuncSigs = map[string]string{
	"9f46e9d7": "checkSignature(address,bytes32,uint8,bytes32,bytes32)",
}

// SigningTestBin is the compiled bytecode used for deploying new contracts.
var SigningTestBin = "0x608060405234801561001057600080fd5b506101fe806100206000396000f3fe608060405234801561001057600080fd5b506004361061002b5760003560e01c80639f46e9d714610030575b600080fd5b61004361003e366004610167565b610045565b005b6040805160208101829052601c60608201527f19457468657265756d205369676e6564204d6573736167653a0a333200000000608082015290810185905260009060a00160408051601f1981840301815282825280516020918201206000845290830180835281905260ff8716918301919091526060820185905260808201849052915060019060a0016020604051602081039080840390855afa1580156100f1573d6000803e3d6000fd5b505050602060405103516001600160a01b0316866001600160a01b03161461015f5760405162461bcd60e51b815260206004820152601960248201527f5369676e617475726520646f6573206e6f74206d617463682e00000000000000604482015260640160405180910390fd5b505050505050565b600080600080600060a0868803121561017e578081fd5b85356001600160a01b0381168114610194578182fd5b945060208601359350604086013560ff811681146101b0578182fd5b9497939650939460608101359450608001359291505056fea2646970667358221220195e95e03635af188029fdf77ab01bb0aae6ca87b3f0cc533a6de75fb70b88e564736f6c63430008020033"

// DeploySigningTest deploys a new Ethereum contract, binding an instance of SigningTest to it.
func DeploySigningTest(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *SigningTest, error) {
	parsed, err := abi.JSON(strings.NewReader(SigningTestABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}

	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(SigningTestBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &SigningTest{SigningTestCaller: SigningTestCaller{contract: contract}, SigningTestTransactor: SigningTestTransactor{contract: contract}, SigningTestFilterer: SigningTestFilterer{contract: contract}}, nil
}

// SigningTest is an auto generated Go binding around an Ethereum contract.
type SigningTest struct {
	SigningTestCaller     // Read-only binding to the contract
	SigningTestTransactor // Write-only binding to the contract
	SigningTestFilterer   // Log filterer for contract events
}

// SigningTestCaller is an auto generated read-only Go binding around an Ethereum contract.
type SigningTestCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SigningTestTransactor is an auto generated write-only Go binding around an Ethereum contract.
type SigningTestTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SigningTestFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type SigningTestFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// SigningTestSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type SigningTestSession struct {
	Contract     *SigningTest      // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// SigningTestCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type SigningTestCallerSession struct {
	Contract *SigningTestCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts      // Call options to use throughout this session
}

// SigningTestTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type SigningTestTransactorSession struct {
	Contract     *SigningTestTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts      // Transaction auth options to use throughout this session
}

// SigningTestRaw is an auto generated low-level Go binding around an Ethereum contract.
type SigningTestRaw struct {
	Contract *SigningTest // Generic contract binding to access the raw methods on
}

// SigningTestCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type SigningTestCallerRaw struct {
	Contract *SigningTestCaller // Generic read-only contract binding to access the raw methods on
}

// SigningTestTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type SigningTestTransactorRaw struct {
	Contract *SigningTestTransactor // Generic write-only contract binding to access the raw methods on
}

// NewSigningTest creates a new instance of SigningTest, bound to a specific deployed contract.
func NewSigningTest(address common.Address, backend bind.ContractBackend) (*SigningTest, error) {
	contract, err := bindSigningTest(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &SigningTest{SigningTestCaller: SigningTestCaller{contract: contract}, SigningTestTransactor: SigningTestTransactor{contract: contract}, SigningTestFilterer: SigningTestFilterer{contract: contract}}, nil
}

// NewSigningTestCaller creates a new read-only instance of SigningTest, bound to a specific deployed contract.
func NewSigningTestCaller(address common.Address, caller bind.ContractCaller) (*SigningTestCaller, error) {
	contract, err := bindSigningTest(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &SigningTestCaller{contract: contract}, nil
}

// NewSigningTestTransactor creates a new write-only instance of SigningTest, bound to a specific deployed contract.
func NewSigningTestTransactor(address common.Address, transactor bind.ContractTransactor) (*SigningTestTransactor, error) {
	contract, err := bindSigningTest(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &SigningTestTransactor{contract: contract}, nil
}

// NewSigningTestFilterer creates a new log filterer instance of SigningTest, bound to a specific deployed contract.
func NewSigningTestFilterer(address common.Address, filterer bind.ContractFilterer) (*SigningTestFilterer, error) {
	contract, err := bindSigningTest(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &SigningTestFilterer{contract: contract}, nil
}

// bindSigningTest binds a generic wrapper to an already deployed contract.
func bindSigningTest(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(SigningTestABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SigningTest *SigningTestRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SigningTest.Contract.SigningTestCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SigningTest *SigningTestRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SigningTest.Contract.SigningTestTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SigningTest *SigningTestRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SigningTest.Contract.SigningTestTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_SigningTest *SigningTestCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _SigningTest.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_SigningTest *SigningTestTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _SigningTest.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_SigningTest *SigningTestTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _SigningTest.Contract.contract.Transact(opts, method, params...)
}

// CheckSignature is a free data retrieval call binding the contract method 0x9f46e9d7.
//
// Solidity: function checkSignature(address _signer, bytes32 _theHash, uint8 _v, bytes32 _r, bytes32 _s) view returns()
func (_SigningTest *SigningTestCaller) CheckSignature(opts *bind.CallOpts, _signer common.Address, _theHash [32]byte, _v uint8, _r [32]byte, _s [32]byte) error {
	var out []interface{}
	err := _SigningTest.contract.Call(opts, &out, "checkSignature", _signer, _theHash, _v, _r, _s)

	if err != nil {
		return err
	}

	return err

}

// CheckSignature is a free data retrieval call binding the contract method 0x9f46e9d7.
//
// Solidity: function checkSignature(address _signer, bytes32 _theHash, uint8 _v, bytes32 _r, bytes32 _s) view returns()
func (_SigningTest *SigningTestSession) CheckSignature(_signer common.Address, _theHash [32]byte, _v uint8, _r [32]byte, _s [32]byte) error {
	return _SigningTest.Contract.CheckSignature(&_SigningTest.CallOpts, _signer, _theHash, _v, _r, _s)
}

// CheckSignature is a free data retrieval call binding the contract method 0x9f46e9d7.
//
// Solidity: function checkSignature(address _signer, bytes32 _theHash, uint8 _v, bytes32 _r, bytes32 _s) view returns()
func (_SigningTest *SigningTestCallerSession) CheckSignature(_signer common.Address, _theHash [32]byte, _v uint8, _r [32]byte, _s [32]byte) error {
	return _SigningTest.Contract.CheckSignature(&_SigningTest.CallOpts, _signer, _theHash, _v, _r, _s)
}
