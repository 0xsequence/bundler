.PHONY: all

all:
	forge build --via-ir --evm-version paris --optimize --optimizer-runs 1
	mkdir -p gen/solabis
	(rm -r gen/solabis/* || true)
	mkdir -p gen/solabis/abivalidator
	mkdir -p gen/solabis/abiendorser
	mkdir -p gen/solabis/abimockendorser
	mkdir -p gen/solabis/abimockwallet
	mkdir -p gen/solabis/abierc20
	node patch-calls.js
	jq .abi < out/OperationValidator.sol/OperationValidator.json | abigen -abi - --pkg abivalidator --type OperationValidator --out gen/solabis/abivalidator/abivalidator.go
	jq .abi < out/Endorser.sol/Endorser.json | abigen -abi - --pkg abiendorser --type Endorser --out gen/solabis/abiendorser/abiendorser.go
	jq .abi < out/MockEndorser.sol/MockEndorser.json | abigen -abi - --pkg abimockendorser --type MockEndorser --out gen/solabis/abimockendorser/abimockendorser.go
	jq .abi < out/MockWallet.sol/MockWallet.json | abigen -abi - --pkg abimockwallet --type MockWallet --out gen/solabis/abimockwallet/abimockwallet.go
	jq .abi < out/ERC20.sol/ERC20.json | abigen -abi - --pkg abierc20 --type ERC20 --out gen/solabis/abierc20/abierc20.go
	find . -type f -name "*.go" -exec sed -i 's|ethereum/go-ethereum|0xsequence/ethkit/go-ethereum|g' {} +
