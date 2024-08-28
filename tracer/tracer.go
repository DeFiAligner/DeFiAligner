package tracer

import (
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/DeFiAligner/DeFiAligner/sevm"
	"github.com/aclements/go-z3/z3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gookit/color"
)

var RPC_URL = sevm.RPC_URL // loacl

type TransactionTrace struct {
	Gas         uint64 `json:"gas"`
	ReturnValue string `json:"returnValue"`
	StructLogs  []struct {
		Pc      uint64            `json:"pc"`
		Op      string            `json:"op"`
		Gas     uint64            `json:"gas"`
		GasCost uint64            `json:"gasCost"`
		Memory  []string          `json:"memory"`
		Stack   []string          `json:"stack"`
		Storage map[string]string `json:"storage"` // 嵌套映射
		Depth   int               `json:"depth"`
		Error   string            `json:"error,omitempty"`
	} `json:"structLogs"`
}

func GetTransactionTrace(txHash common.Hash) *TransactionTrace {
	ethClient, err := ethclient.Dial(
		RPC_URL,
	)
	if err != nil {
		panic(err)
	}

	var result TransactionTrace
	ethClient.Client().Call(
		&result,
		"debug_traceTransaction",
		txHash,
		map[string]interface{}{
			"disableStack":   false,
			"disableMemory":  false,
			"disableStorage": false,
			"onlyTopCall":    false,
		},
	)

	return &result

}

func (trace TransactionTrace) PrintTrace() {

	for i, log := range trace.StructLogs {

		if i < len(trace.StructLogs)-1 {

			fmt.Println("============================================")
			fmt.Println("        *****PC******:", log.Pc)
			fmt.Println("        *****OP******:", log.Op)

			fmt.Printf("\tMemory: %v\n", trace.StructLogs[i+1].Memory)
			fmt.Printf("\tStack: %v\n", trace.StructLogs[i+1].Stack)

			if log.Error != "" {
				fmt.Printf("\tError: %s\n", log.Error)
			}
			// Print the storage map
			fmt.Printf("\tStorage:\n")
			for key, value := range log.Storage {
				fmt.Printf("\t\t%s: %s\n", key, value)
			}
			fmt.Println()

		}

	}

}

func HexToBigInt(hexStr string) *big.Int {
	// Remove the 0x prefix if present.
	if len(hexStr) > 2 && hexStr[0:2] == "0x" {
		hexStr = hexStr[2:]
	}
	// Create a big.Int and set it's value from the hex string.
	n := new(big.Int)
	n, _ = n.SetString(hexStr, 16) // 16 for hexadecimal
	return n
}

func HexToBool(hexStr string) bool {
	// Convert the hexadecimal string to an integer.
	value, err := strconv.ParseUint(hexStr, 0, 64)
	if err != nil {
		return false
	}

	// In Go, any non-zero value is true, and 0 is false.
	return value != 0
}

func CompareStackValue(stack_value1 string, stack_value2 z3.Value) bool {
	bv_type := reflect.TypeOf(stack_value2.Context().BVConst("", 256))
	int_type := reflect.TypeOf(stack_value2.Context().IntConst(""))
	bool_type := reflect.TypeOf(stack_value2.Context().BoolConst(""))
	stack_value2 = stack_value2.Context().Simplify(stack_value2, stack_value2.Context().Config())

	stack_bigInt_value1 := HexToBigInt(stack_value1)

	// bv type
	if reflect.TypeOf(stack_value2) == bv_type {
		stack_bigInt_value2, ok := stack_value2.(z3.BV).AsBigUnsigned()
		if !ok {
			return true
		} else if stack_bigInt_value1.Cmp(stack_bigInt_value2) == 0 {
			return true
		}

	}

	// int type
	if reflect.TypeOf(stack_value2) == int_type {
		stack_bigInt_value2, ok := stack_value2.(z3.Int).AsBigInt()
		if !ok {
			return true
		} else if stack_bigInt_value1.Cmp(stack_bigInt_value2) == 0 {
			return true
		}

	}

	// boo type
	if reflect.TypeOf(stack_value2) == bool_type {
		stack_bool_value2, ok := stack_value2.(z3.Bool).AsBool()
		if !ok {
			return true
		} else if stack_bool_value2 == HexToBool(stack_value1) {
			return true
		}

	}

	return false
}

func CompareMomoryValue(symbol_memory sevm.SymbolicMemory, real_memory []string) bool {
	real_memory_length := len(real_memory)
	symbol_memory_length := symbol_memory.Store.Sort().BVSize()
	for i := 0; i < real_memory_length; i++ {

		if i*256+256 <= symbol_memory_length {
			symbol_value := sevm.SimplifyZ3BV(symbol_memory.Store.(z3.BV).Extract(i*256+256-1, i*256))
			symbol_value_big, ok := symbol_value.AsBigUnsigned()
			if ok {
				real_value_bigint := new(big.Int)
				real_value_bigint.SetString(real_memory[i], 16)
				if symbol_value_big.Cmp(real_value_bigint) != 0 {
					color.Error.Println("Different Index:", i, symbol_value, real_memory[i])
					return false
				}
			}

		}

	}
	return true //same

}

func Print_ValueArray(memory []string) {
	for index, value := range memory {
		formattedStr := fmt.Sprintf("Slot: %-4d Offset: %08X Size: %-4d Value: %s", index, index*32, 32, value)
		color.Yellow.Println(formattedStr)
	}

}

func CompareTraces(tx_hash string, interpreter sevm.SymbolicEVMInterpreter, contract *sevm.Contract, ctx *z3.Context) bool {

	/// 01. Get trace from blockchain
	historical_trace := GetTransactionTrace(
		common.HexToHash(tx_hash),
	)

	var executionPathList = make(sevm.ExecutionPathList, 0)
	var currentPath = make(sevm.ExecutionPath, 0) // current opcode Path

	interpreter.SymbolicRun_DFS(contract, ctx, currentPath, &executionPathList)

	color.Error.Println("The number of Pathes", len(executionPathList))

	// 02. Compare traces
	for index, path := range executionPathList {
		if CompareResults(historical_trace, path) {
			color.Yellow.Println("The index of the same path is:", index)
			return true
		}
	}

	color.Yellow.Println("All paths are different", "total: ", len(executionPathList))
	return false

}

func CompareResults(historical_trace *TransactionTrace, evmExecutionStateList sevm.ExecutionPath) bool {
	color.Infoln("*********************Start Compare Traces: ****************")

	for index, historical_state := range historical_trace.StructLogs {

		if index < len(historical_trace.StructLogs)-1 {

			// history

			historical_pc := historical_state.Pc
			historical_opcode := historical_state.Op

			// secm
			sevm_pc := evmExecutionStateList[index].Current_pc
			sevm_opcode := evmExecutionStateList[index].Current_opcode

			sevm_value := evmExecutionStateList[index].Current_return_value

			sevm_error := evmExecutionStateList[index].Current_return_error

			color.Blue.Println("===Debug====", historical_pc, historical_opcode, sevm_pc, sevm_opcode, sevm_value, sevm_error)

			evmExecutionStateList[index].Current_stack.PrintCurrentStack()
			evmExecutionStateList[index].Current_memory.PrintSymbolicMemory()
			fmt.Println(historical_trace.StructLogs[index+1].Stack)
			Print_ValueArray(historical_trace.StructLogs[index+1].Memory)

			// compare opcode and pc
			if historical_pc != sevm_pc || sevm_opcode.String() != historical_opcode {
				fmt.Println("The pc is different: ")
				fmt.Println(historical_pc, sevm_pc)
				fmt.Println(historical_opcode, sevm_opcode.String())

				return false

			}

			// compare stack length
			if len(historical_trace.StructLogs[index+1].Stack) != evmExecutionStateList[index].Current_stack.Len() {
				fmt.Println("The stack length is different:", len(historical_trace.StructLogs[index+1].Stack), evmExecutionStateList[index].Current_stack.Len())

				return false
			}
			// compare stack values
			// fmt.Println("***************************************************Stack****************************************************")
			stack_len := len(historical_trace.StructLogs[index+1].Stack)
			for stack_index := 0; stack_index < stack_len; stack_index++ {
				//fmt.Println(historical_pc, sevm_pc, stack_index, stack_len)

				historical_value := historical_trace.StructLogs[index+1].Stack[stack_len-1-stack_index]
				sevm_value := evmExecutionStateList[index].Current_stack.Back(stack_index)
				if !CompareStackValue(historical_value, sevm_value) {
					fmt.Println("The stack value is different:")
					fmt.Println(historical_pc, sevm_pc)
					fmt.Println(historical_opcode, sevm_opcode.String())
					fmt.Println(historical_value, sevm_value)
					return false
				}

			}

			//compare memory values
			if !CompareMomoryValue(evmExecutionStateList[index].Current_memory, historical_trace.StructLogs[index+1].Memory) {

				fmt.Println("The memory value is different:")

				return false

			}

		}
	}
	color.Infof("Find the same path > > > > ")
	return true
}
func CompareResultsWithoutPrint(historical_trace *TransactionTrace, evmExecutionStateList sevm.ExecutionPath) bool {
	color.Infoln("*********************Start Compare Traces: ****************")

	for index, historical_state := range historical_trace.StructLogs {

		if index < len(historical_trace.StructLogs)-1 {

			// history

			historical_pc := historical_state.Pc
			historical_opcode := historical_state.Op

			// secm
			sevm_pc := evmExecutionStateList[index].Current_pc
			sevm_opcode := evmExecutionStateList[index].Current_opcode

			// sevm_value := evmExecutionStateList[index].Current_return_value

			// sevm_error := evmExecutionStateList[index].Current_return_error

			//color.Blue.Println("=====================================", historical_pc, historical_opcode, sevm_pc, sevm_opcode, sevm_value, sevm_error, "=====================================")

			// evmExecutionStateList[index].Current_stack.PrintCurrentStack()
			// evmExecutionStateList[index].Current_memory.PrintSymbolicMemory()
			// fmt.Println(historical_trace.StructLogs[index+1].Stack)
			// Print_ValueArray(historical_trace.StructLogs[index+1].Memory)

			// compare opcode and pc
			if historical_pc != sevm_pc || sevm_opcode.String() != historical_opcode {
				fmt.Println("The pc is different: ")
				fmt.Println(historical_pc, sevm_pc)
				fmt.Println(historical_opcode, sevm_opcode.String())

				return false

			}

			// compare stack length
			if len(historical_trace.StructLogs[index+1].Stack) != evmExecutionStateList[index].Current_stack.Len() {
				fmt.Println("The stack length is different:", len(historical_trace.StructLogs[index+1].Stack), evmExecutionStateList[index].Current_stack.Len())

				return false
			}
			// compare stack values
			// fmt.Println("***************************************************Stack****************************************************")
			stack_len := len(historical_trace.StructLogs[index+1].Stack)
			for stack_index := 0; stack_index < stack_len; stack_index++ {
				//fmt.Println(historical_pc, sevm_pc, stack_index, stack_len)

				historical_value := historical_trace.StructLogs[index+1].Stack[stack_len-1-stack_index]
				sevm_value := evmExecutionStateList[index].Current_stack.Back(stack_index)
				if !CompareStackValue(historical_value, sevm_value) {
					// fmt.Println("The stack value is different:")
					// fmt.Println(historical_pc, sevm_pc)
					// fmt.Println(historical_opcode, sevm_opcode.String())
					// fmt.Println(historical_value, sevm_value)
					return false
				}

			}

			//compare memory values
			if !CompareMomoryValue(evmExecutionStateList[index].Current_memory, historical_trace.StructLogs[index+1].Memory) {

				//fmt.Println("The memory value is different:")

				return false

			}

		}
	}
	color.Infof("Find the same path > > > > ")
	return true
}
