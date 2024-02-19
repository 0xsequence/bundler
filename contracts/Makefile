.PHONY: all

all:
	forge build --via-ir
	mkdir -p gen/operationvalidator && jq .abi < out/OperationValidator.sol/OperationValidator.json | abigen -abi - --pkg operationvalidator --type OperationValidator --out gen/operationvalidator/operationvalidator.go
	fd '\.go' | xargs sd ethereum/go-ethereum 0xsequence/ethkit/go-ethereum
