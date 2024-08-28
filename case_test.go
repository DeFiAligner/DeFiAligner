package test_case

import (
	"fmt"
	"github.com/DeFiAligner/DeFiAligner/abiparser"
	"github.com/DeFiAligner/DeFiAligner/pathfeat"
	"github.com/DeFiAligner/DeFiAligner/sevm"
	"github.com/aclements/go-z3/z3"
	"github.com/ethereum/go-ethereum/common"
	"testing"
)

func TestCase(t *testing.T) {

	// Function input
	filePath := "./abiparser/ABIs/erc20.abi.json"
	abiElements, err := abiparser.ReadAndParseABI(filePath)
	if err != nil {
		t.Fatalf("ReadAndParseABI() failed with error: %s", err)
	}
	element := abiparser.SelectABIFunction(abiElements, "transfer")
	fmt.Printf("Type: %s, Name: %s\n", element.Type, element.Name)
	ctx := z3.NewContext(nil)
	specified_values := make(map[string]z3.BV)

	symbolic_inputData, intput_Z3Variables := abiparser.GenerateSymbolicInput(element, ctx, specified_values)
	fmt.Println(symbolic_inputData)
	fmt.Println(intput_Z3Variables)

	// Txs Info
	to_address := common.HexToAddress("0x4306B12F8e824cE1fa9604BbD88f2AD4f0FE3c54").Big()
	contractRef := ctx.FromBigInt(to_address, ctx.BVSort(256)).(z3.BV)
	caller := ctx.BVConst("Caller", 256)
	value := ctx.BVConst("ETHValue", 256)

	contract := sevm.NewContract(caller, contractRef, value, 0) // to code
	address_string := "0x" + contractRef.String()[26:66]
	contract.Code = sevm.GetContractCodeByAddress(address_string)

	// Initialization
	contract.Input = symbolic_inputData
	interpreter := sevm.NewEVMInterpreter(&sevm.SEVM{})
	var executionPathList = make(sevm.ExecutionPathList, 0)
	var currentPath = make(sevm.ExecutionPath, 0) // current Path
	interpreter.SymbolicRun_DFS(contract, ctx, currentPath, &executionPathList)

	// update Variables
	interpreter.Z3Variables["Caller"] = caller
	interpreter.Z3Variables["ETHValue"] = value
	interpreter.Z3Variables["Address_to"] = contractRef

	fmt.Println("There are ", len(executionPathList), "paths: ")

	for _, path := range executionPathList {
		features := pathfeat.GetDeFiSymbolicFeaturesFromPath(path, contractRef)
		if len(features.ERCBalanceChangeMap) > 0 {
			fmt.Println(features.ERCBalanceChangeMap)
			for _, list := range features.CondictionMap {
				for _, con := range list {
					fmt.Println(pathfeat.SimplifySymbolLogic(con.String()))
				}

			}

		}

	}

}
