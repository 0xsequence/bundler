
const fs = require('fs')

function main() {
  // Read out/OperationValidator.sol/OperationValidator.json
  const operationValidator = fs.readFileSync('out/OperationValidator.sol/OperationValidator.json', 'utf8')

  // In the ABI, find the type=function and name=simulateOperation
  const obj = JSON.parse(operationValidator)
  const simulateOperation = obj.abi.find((x) => x.type === 'function' && x.name === 'simulateOperation')

  // Change the stateMutability to view
  simulateOperation.stateMutability = 'view'

  // Write the ABI back to out/OperationValidator.sol/OperationValidator.json
  fs.writeFileSync('out/OperationValidator.sol/OperationValidator.json', JSON.stringify(obj, null, 2))
}

main()
