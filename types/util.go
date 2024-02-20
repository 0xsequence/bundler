package types

import (
	"fmt"
	"math/big"

	"github.com/0xsequence/ethkit/go-ethereum/common/hexutil"
)

func FromHex(hex string) ([]byte, error) {
	return hexutil.Decode(hex)
}

func HexToBigInt(str string) (*big.Int, error) {
	var ok bool
	var res *big.Int

	// If starts with 0x then it is a hex string
	if len(str) >= 2 && str[:2] == "0x" {
		res, ok = new(big.Int).SetString(str[2:], 16)
	} else {
		res, ok = new(big.Int).SetString(str, 10)
	}

	if !ok {
		return nil, fmt.Errorf("invalid big int string: %s", str)
	}

	return res, nil
}
