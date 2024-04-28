package pricefeed

import (
	"github.com/0xsequence/bundler/lib/pricefeed/abis"
	"github.com/0xsequence/ethkit/ethcontract"
	"github.com/0xsequence/ethkit/ethrpc"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

func FetchDecimals(provider ethrpc.Interface, token common.Address) (int, error) {
	abi := ethcontract.MustParseABI(abis.ERC20)
	contract := ethcontract.NewContractCaller(token, abi, provider)

	var result []interface{}
	err := contract.Call(nil, &result, "decimals")
	if err != nil {
		return 0, err
	}

	return int(result[0].(uint8)), nil
}
