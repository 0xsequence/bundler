.PHONY: all

all:
	cp src/OperationValidator.sol src/OperationValidatorSimulator.sol
	patch src/OperationValidatorSimulator.sol simulateOperation.patch
	forge build --via-ir
	mkdir -p gen/operationvalidator
	jq .abi < out/OperationValidator.sol/OperationValidator.json | abigen -abi - --pkg operationvalidator --type OperationValidator --out gen/operationvalidator/operationvalidator.go
	jq .abi < out/OperationValidatorSimulator.sol/OperationValidatorSimulator.json | abigen -abi - --pkg operationvalidator --type OperationValidatorSimulator --out gen/operationvalidator/operationvalidatorsimulator.go
	fd '\.go' | xargs sd ethereum/go-ethereum 0xsequence/ethkit/go-ethereum
