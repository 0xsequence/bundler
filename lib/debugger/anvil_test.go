package debugger_test

import (
	"context"
	"os/exec"
	"testing"

	"github.com/0xsequence/bundler/lib/debugger"
	"github.com/0xsequence/ethkit/go-ethereum/common"
	"github.com/go-chi/httplog/v2"
	"github.com/stretchr/testify/assert"
)

func TestDebugWithOverrides(t *testing.T) {
	// Test set up
	ctx := context.Background()
	level := httplog.LevelByName("DEBUG")
	logger := httplog.NewLogger("", httplog.Options{
		LogLevel: level,
	})
	rpcUrl := "http://localhost:8545" // Set below

	anvil, err := debugger.NewAnvilDebugger(ctx, logger, nil, rpcUrl)
	if err != nil {
		// Anvil not available. Skip test
		t.Skip(err)
	}

	// Run new anvil instance as the RPC to clone
	cmd := exec.Command("anvil")
	err = cmd.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		cmd.Process.Kill()
	}()

	codeAddr := common.HexToAddress("0x1234567890123456789012345678901234567890")
	// MockERC20
	code := "0x6080604052348015600f57600080fd5b506004361060285760003560e01c806370a0823114602d575b600080fd5b605360383660046065565b6001600160a01b031660009081526020819052604090205490565b60405190815260200160405180910390f35b600060208284031215607657600080fd5b81356001600160a01b0381168114608c57600080fd5b939250505056fea26469706673582212204a2495d07316942a5b44bf5c22ecae8e15845ab84942d1ba514dfd2afdf6d1ba64736f6c63430008180033"
	// keccak256(abi.encode(address(0), bytes32(0)))
	slot := common.HexToHash("0xad3228b676f7d3cd4284a5443f17f1962b36e491b30a40b2405849e597ba5fb5")
	slotValue := common.HexToHash("0x100")

	// MockERC20.balanceOf(address(0))
	encodedCallData := "0x70a082310000000000000000000000000000000000000000000000000000000000000000"
	args := &debugger.DebugCallArgs{
		From: common.Address{},
		To:   codeAddr,
		Data: common.FromHex(encodedCallData),
	}

	overrideArgs := &debugger.DebugOverrideArgs{
		codeAddr: {
			Code: &code,
			StateDiff: map[common.Hash]common.Hash{
				slot: slotValue,
			},
		},
	}

	// Call it
	result, err := anvil.DebugTraceCall(ctx, args, overrideArgs)

	// Assert result
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	assert.False(t, result.Failed)
	assert.Equal(t, slotValue.String(), "0x"+result.ReturnValue)
}
