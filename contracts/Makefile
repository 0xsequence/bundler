.PHONY: all

all:
	forge build --via-ir
	mkdir -p gen/operationvalidator && jq .abi < out/OperationValidator.sol/OperationValidator.json | abigen -abi - --pkg operationvalidator --out gen/operationvalidator/operationvalidator.go
