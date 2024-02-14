package proto

import (
	"encoding/binary"
	"math/big"

	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
	"github.com/0xsequence/ethkit/go-ethereum/crypto"
)

func (o *Operation) Digest() common.Hash {
	var gasLimit [8]byte
	var endorserGasLimit [8]byte

	binary.LittleEndian.PutUint64(gasLimit[:], o.GasLimit)
	binary.LittleEndian.PutUint64(endorserGasLimit[:], o.EndorserGasLimit)

	maxFeePerGas, ok := new(big.Int).SetString(o.MaxFeePerGas, 0)
	if !ok {
		maxFeePerGas = big.NewInt(0)
	}

	priorityFeePerGas, ok := new(big.Int).SetString(o.PriorityFeePerGas, 0)
	if !ok {
		priorityFeePerGas = big.NewInt(0)
	}

	return crypto.Keccak256Hash(
		common.HexToAddress(o.Entrypoint).Bytes(),
		hexutil.MustDecode(o.CallData),
		gasLimit[:],
		common.HexToAddress(o.FeeToken).Bytes(),
		common.HexToAddress(o.Endorser).Bytes(),
		hexutil.MustDecode(o.EndorserCallData),
		endorserGasLimit[:],
		maxFeePerGas.Bytes(),
		priorityFeePerGas.Bytes(),
	)
}
