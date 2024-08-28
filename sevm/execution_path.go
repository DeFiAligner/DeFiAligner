package sevm

import (
	"github.com/aclements/go-z3/z3"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/gookit/color"
)

type EVMExecutionState struct { //record stack states, memory states and  otther info
	Current_opcode          vm.OpCode
	Current_pc              uint64
	Current_memory          SymbolicMemory
	Current_stack           *SymbolicStack
	Current_return_value    z3.Value
	Current_return_error    error
	Current_evm_depth       int
	Current_called_contract z3.BV
}

type ExecutionPath []EVMExecutionState
type ExecutionPathList []ExecutionPath

func updateExecutionPath(executionPath *ExecutionPath, stack *SymbolicStack, memory *SymbolicMemory, pc uint64, opcode vm.OpCode, ret z3.Value, return_error error, evm_depth int, current_called_contract z3.BV) {

	newState := EVMExecutionState{
		Current_opcode:          opcode,
		Current_pc:              pc,
		Current_memory:          memory.Copy(),
		Current_stack:           stack.Copy(),
		Current_return_value:    ret,
		Current_return_error:    return_error,
		Current_evm_depth:       evm_depth,
		Current_called_contract: current_called_contract,
	}

	*executionPath = append(*executionPath, newState)

}

func DeepCopyExecutionPath(original ExecutionPath) ExecutionPath { //In Golang, all slices are references by default. So, we need a deep copy
	copyPath := make(ExecutionPath, len(original))
	for i, state := range original {
		copyPath[i] = EVMExecutionState{
			Current_opcode:          state.Current_opcode,
			Current_pc:              state.Current_pc,
			Current_memory:          state.Current_memory.Copy(),
			Current_stack:           state.Current_stack.Copy(),
			Current_return_value:    state.Current_return_value,
			Current_return_error:    state.Current_return_error,
			Current_evm_depth:       state.Current_evm_depth,
			Current_called_contract: state.Current_called_contract,
		}
	}
	return copyPath
}

func (list *ExecutionPathList) AddPath(currentPath ExecutionPath) {
	copyPath := DeepCopyExecutionPath(currentPath)
	*list = append(*list, copyPath)
}

func (path_list ExecutionPathList) ExecutionPathListFilter() {

	for _, path := range path_list {
		for index, state := range path {

			if state.Current_opcode == vm.JUMPI {
				color.Println(state.Current_opcode, state.Current_pc)
				color.Infoln(path[index-1].Current_stack.Back(1))
			}
			if state.Current_opcode == vm.SSTORE {
				color.Println(state.Current_opcode, state.Current_pc)
				color.Infoln(path[index-1].Current_stack.Back(1))
			}

		}

		color.Warnln("****************")
	}

}
