.PHONY: all

all:
	forge build --via-ir --evm-version paris --optimize --optimizer-runs 1
	mkdir -p gen/operationvalidator
	rm -r gen/operationvalidator/*
	node patch-simulate.js
	jq .abi < out/OperationValidator.sol/OperationValidator.json | abigen -abi - --pkg operationvalidator --type OperationValidator --out gen/operationvalidator/operationvalidator.go
	find . -type f -name "*.go" -exec sed -i 's|ethereum/go-ethereum|0xsequence/ethkit/go-ethereum|g' {} +

