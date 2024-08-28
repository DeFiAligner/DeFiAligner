package pathfeat

import (
	"github.com/DeFiAligner/DeFiAligner/sevm"
	"github.com/aclements/go-z3/z3"
	"github.com/ethereum/go-ethereum/common"
)

// get the symbolic expression of the user's balance
func GetBalanceSymbol(token_address string, ctx *z3.Context) []z3.BV {
	//0x70a08231
	user_address := ctx.BVConst("+++", 256) // Caller
	address := common.HexToAddress(token_address).Big()
	contractRef := ctx.FromBigInt(address, ctx.BVSort(256)).(z3.BV)
	caller := ctx.BVConst("Caller", 256) // Caller
	value := ctx.BVConst("Value", 256)   // ETH value
	contract := sevm.NewContract(caller, contractRef, value, 0)
	address_string := "0x" + contractRef.String()[26:66]
	contract.Code = sevm.GetContractCodeByAddress(address_string)         // to code
	func_name := ctx.FromInt(1889567281, ctx.IntSort()).(z3.Int).ToBV(32) // 0x70a08231, balanceOf(address)
	inputData := func_name.Concat(user_address)

	contract.Input = inputData
	interpreter := sevm.NewEVMInterpreter(&sevm.SEVM{})
	var executionPathList = make(sevm.ExecutionPathList, 0)
	var currentPath = make(sevm.ExecutionPath, 0) // current Path
	interpreter.SymbolicRun_DFS(contract, ctx, currentPath, &executionPathList)

	return_value_list := []z3.BV{}
	for _, path := range executionPathList {
		if path[len(path)-1].Current_return_value != nil {
			return_value_list = append(return_value_list, path[len(path)-1].Current_return_value.(z3.BV)) // may involve multiple balances.
		}

	}

	return return_value_list

}
