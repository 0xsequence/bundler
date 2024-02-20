.PHONY: all

all:
	forge build --via-ir --evm-version paris --optimize --optimizer-runs 1
	mkdir -p gen/solabis
	(rm -r gen/solabis/* || true)
	mkdir -p gen/solabis/abivalidator
	mkdir -p gen/solabis/abiendorser
	node patch-calls.js
	jq .abi < out/OperationValidator.sol/OperationValidator.json | abigen -abi - --pkg abivalidator --type OperationValidator --out gen/solabis/abivalidator/abivalidator.go
	jq .abi < out/Endorser.sol/Endorser.json | abigen -abi - --pkg abiendorser --type Endorser --out gen/solabis/abiendorser/abiendorser.go
	find . -type f -name "*.go" -exec sed -i 's|ethereum/go-ethereum|0xsequence/ethkit/go-ethereum|g' {} +
