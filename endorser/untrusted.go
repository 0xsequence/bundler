package endorser

import (
	"fmt"

	"github.com/0xsequence/bundler/lib/debugger"
	"github.com/0xsequence/ethkit/go-ethereum/common"
)

// event UntrustedStarted();
const UNTRUSTED_STARTED_SIG = "0x5e802b414aa8d12c8eb59955acbf0f9ce42afc2842ee01c46d7053c94528687a"

// event UntrustedEnded();
const UNTRUSTED_ENDED_SIG = "0xe59e021ea70d7129da268f7f72da44741fc74a90aa5dde2db22588e4c4e3e34d"

func isZeros(b []byte) bool {
	for _, v := range b {
		if v != 0 {
			return false
		}
	}
	return true
}

// TODO: Handle CREATE and CREATE2
func ParseUntrustedDebug(tt *debugger.TransactionTrace) (*EndorserResult, error) {
	result := &EndorserResult{}

	selfstack := make([]common.Address, 1)
	selfstack[0] = tt.From

	untrustedDepth := 0
	var depthOfUntrusted *float64

	// Iterate over every opcode
	for i, log := range tt.StructLogs {
		switch log.Op {
		case "CALL", "STATICCALL":
			// Change the self address
			l := len(log.Stack)
			if l >= 2 {
				selfstack = append(selfstack, common.HexToAddress(log.Stack[l-2]))
			}

		case "CALLDCODE", "DELEGATECALL":
			// Change the self address (to self again)
			selfstack = append(selfstack, selfstack[len(selfstack)-1])

		case "RETURN", "REVERT":
			// Change the self address
			if len(selfstack) == 0 {
				return nil, fmt.Errorf("selfstack is empty")
			}

			selfstack = selfstack[:len(selfstack)-1]

		case "LOG1":
			// The execution may be entering or leaving the untrusted code
			// LOG1 opcode takes offset, size, topic
			// we only care about the topic being one of the two
			l := len(log.Stack)
			if l >= 3 {
				cand := log.Stack[l-3]
				if cand == UNTRUSTED_STARTED_SIG {
					untrustedDepth++
					if depthOfUntrusted == nil {
						depthOfUntrusted = &log.Depth
					}
				} else if cand == UNTRUSTED_ENDED_SIG {
					// Ignore if we aren't in the right depth
					if depthOfUntrusted != nil && log.Depth == *depthOfUntrusted {
						untrustedDepth--
						if untrustedDepth == 0 {
							depthOfUntrusted = nil
						}
					}
				}
			}

		case "BALANCE":
			if untrustedDepth > 0 {
				l := len(log.Stack)
				if l >= 1 {
					addr := common.HexToAddress(log.Stack[l-1])
					result.SetBalance(addr, true)
				}
			}

		case "ORIGIN":
			if untrustedDepth > 0 {
				result.SetOrigin(true)
			}

		case "GASPRICE":
			if untrustedDepth > 0 {
				result.SetGasPrice(true)
			}

		case "EXTCODESIZE", "EXTCODECOPY", "EXTCODEHASH":
			if untrustedDepth > 0 {
				l := len(log.Stack)
				ni := i + 1
				if l >= 1 && len(tt.StructLogs) > ni {
					// The result and the argument use the same position on the stack
					// as the opcodes takes 1 and returns 1
					sres := common.FromHex(tt.StructLogs[ni].Stack[l-1])
					if isZeros(sres) {
						addr := common.HexToAddress(log.Stack[l-1])
						result.SetCode(addr, true)
					}
				}
			}

		case "COINBASE":
			if untrustedDepth > 0 {
				result.SetCoinbase(true)
			}

		case "TIMESTAMP":
			if untrustedDepth > 0 {
				result.SetTimestamp(true)
			}

		case "NUMBER":
			if untrustedDepth > 0 {
				result.SetNumber(true)
			}

		case "DIFFICULTY", "PREVRANDAO":
			if untrustedDepth > 0 {
				result.SetDifficulty(true)
			}

		case "CHAINID":
			if untrustedDepth > 0 {
				result.SetChainID(true)
			}

		case "SELFBALANCE":
			if untrustedDepth > 0 {
				result.SetBalance(selfstack[len(selfstack)-1], true)
			}

		case "BASEFEE":
			if untrustedDepth > 0 {
				result.SetBasefee(true)
			}

		case "SLOAD":
			if untrustedDepth > 0 {
				l := len(log.Stack)
				if l >= 1 {
					result.SetStorageSlot(selfstack[len(selfstack)-1], common.HexToHash(log.Stack[l-1]), true)
				}
			}

		case "CREATE":
			if untrustedDepth > 0 {
				return nil, fmt.Errorf("CREATE is not supported on untrusted code")
			}

		case "CREATE2":
			if untrustedDepth > 0 {
				// Loop through ops to find the address of the created contract set on depth return
				addr := common.Address{}
				createDepth := log.Depth
				for j := i + 1; j < len(tt.StructLogs); j++ {
					if tt.StructLogs[j].Depth == createDepth {
						l := len(tt.StructLogs[j].Stack)
						addr = common.HexToAddress(tt.StructLogs[j].Stack[l-1])
						result.SetCode(addr, true)
						break
					}
				}
				selfstack = append(selfstack, addr)
			}

		default:
		}
	}

	return result, nil
}
