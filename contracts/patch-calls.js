
const fs = require('fs')

function patchABI(abiDir, funcName) {
  const abi = fs.readFileSync(abiDir, 'utf8')

  const obj = JSON.parse(abi)
  const func = obj.abi.find((x) => x.type === 'function' && x.name === funcName)

  if (!func) {
    console.error(`Function ${funcName} not found in ${abiDir}`)
    return
  }

  func.stateMutability = 'view'

  fs.writeFileSync(abiDir, JSON.stringify(obj, null, 2))

}

function main() {
  patchABI('out/Endorser.sol/Endorser.json', 'isOperationReady')
  patchABI('out/OperationValidator.sol/OperationValidator.json', 'simulateOperation')
}

main()
