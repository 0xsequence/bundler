.PHONY: all

all:
	forge build --via-ir
	mkdir -p gen/entrypoint && jq .abi < out/Entrypoint.sol/BundlerEntrypoint.json | abigen -abi - --pkg entrypoint --out gen/entrypoint/entrypoint.go
