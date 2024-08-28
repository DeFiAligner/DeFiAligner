package abiparser

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aclements/go-z3/z3"
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/crypto/sha3"
)

// Define a struct to match the structure of the ABI JSON
type ABIElement struct {
	Constant        bool   `json:"constant"`
	Payable         bool   `json:"payable"`
	StateMutability string `json:"stateMutability"`
	Type            string `json:"type"`
	Name            string `json:"name,omitempty"`
	Inputs          []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"inputs,omitempty"`
	Outputs []struct {
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"outputs,omitempty"`
}

// Function to read and parse the ABI JSON file
func ReadAndParseABI(filePath string) ([]ABIElement, error) {
	// Read the file
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Parse the JSON into the struct
	var abiElements []ABIElement
	if err := json.Unmarshal(fileContent, &abiElements); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return abiElements, nil
}

// Select a specific function in the ABI according the func name
func SelectABIFunction(abiElements []ABIElement, funcName string) ABIElement {
	for _, element := range abiElements {
		if element.Type == "function" && element.Name == funcName { //must a  function
			return element
		}
	}
	return ABIElement{}
}

// Generate Funcation Hash
func GenerateEthereumFunctionHash(abi ABIElement) string {
	inputTypes := make([]string, len(abi.Inputs))
	for i, input := range abi.Inputs {
		inputTypes[i] = input.Type
	}
	signature := fmt.Sprintf("%s(%s)", abi.Name, strings.Join(inputTypes, ","))

	hash := sha3.NewLegacyKeccak256()
	hash.Write([]byte(signature))
	hashBytes := hash.Sum(nil)[:4]        // Correctly obtain the first 4 bytes
	return fmt.Sprintf("0x%x", hashBytes) // Convert the first 4 bytes to hex string
}

// Abi type to Z3
func AbiTypeToZ3Value(abiType string, name string, ctx *z3.Context) z3.BV {
	switch {
	case strings.HasPrefix(abiType, "uint") || strings.HasPrefix(abiType, "int"):
		return ctx.BVConst(name, 256)
	case abiType == "address":
		return ctx.BVConst(name, 256)
	case abiType == "bool":
		return ctx.BVConst(name, 256)
	case strings.HasPrefix(abiType, "bytes") && abiType != "bytes":
		//##  array_loaction
		return ctx.BVConst(name, 256)
	case abiType == "string":
		return ctx.BVConst(name, 256)
	default: // unsolved:  uint256[]  uint256[][]  struct   enum  tuple
		array_loaction_big := common.HexToAddress("0000000000000000000000000000000000000000000000000000000000000080").Big()
		array_loaction := ctx.FromBigInt(array_loaction_big, ctx.BVSort(256)).(z3.BV)
		zero_data := ctx.FromInt(0, ctx.IntSort()).(z3.Int).ToBV(256)
		return array_loaction.Concat(zero_data)
	}
}

// Automatically Generate Symbolic Input Based on ABI
func GenerateSymbolicInput(abiElement ABIElement, ctx *z3.Context, specified_values map[string]z3.BV) (z3.BV, map[string]z3.BV) {

	function_hash_str := GenerateEthereumFunctionHash(abiElement)
	function_hash_decimal, _ := strconv.ParseInt(function_hash_str[2:], 16, 64)

	symbolic_inputData := ctx.FromInt(function_hash_decimal, ctx.IntSort()).(z3.Int).ToBV(32)

	intput_Z3Variables := make(map[string]z3.BV)
	for _, input := range abiElement.Inputs {
		var value z3.BV
		if specifiedVal, exists := specified_values[input.Name]; exists {
			value = specifiedVal
		} else {
			value = AbiTypeToZ3Value(input.Type, input.Name, ctx)
		}
		intput_Z3Variables[input.Name] = value.Context().Simplify(value, ctx.Config()).(z3.BV)
		symbolic_inputData = symbolic_inputData.Concat(value)
	}
	symbolic_inputData = symbolic_inputData.Context().Simplify(symbolic_inputData, ctx.Config()).(z3.BV)
	return symbolic_inputData, intput_Z3Variables
}
