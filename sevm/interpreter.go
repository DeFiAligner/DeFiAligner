package sevm

import (
	"errors"
	// "os"
	"strings"

	"github.com/aclements/go-z3/z3"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/gookit/color"
)

// ScopeContext contains the things that are per-call, such as stack and memory,
type ScopeContext struct {
	Stack    *SymbolicStack
	Contract *Contract
	Memory   *SymbolicMemory
	Z3Contex *z3.Context
}

type SymbolicEVMInterpreter struct {
	evm         *SEVM
	returnData  z3.Value // Last CALL's return data for subsequent reuse
	table       *JumpTable
	Z3Variables map[string]z3.BV //store all global Z3 variables
}

func NewEVMInterpreter(evm *SEVM) *SymbolicEVMInterpreter {
	var table *JumpTable = &symbolicJumpTable
	interpteter := &SymbolicEVMInterpreter{evm: evm, table: table}
	interpteter.evm.interpreter = interpteter
	interpteter.evm.depth = 0
	interpteter.Z3Variables = make(map[string]z3.BV)
	return interpteter
}

type VisitedInfo struct {
	Stacks     []*SymbolicStack //
	Memory     []z3.Value
	ReturnData []z3.Value //n note  used
	Count      []int      // count  for each   state
}

type VisitedNodes map[uint64]map[uint64]*VisitedInfo

const VISIST_Threshold = 1

func isVisitedNode(current_pc uint64, next_pc uint64, stackValue *SymbolicStack, visitedNodes *VisitedNodes, memory z3.Value, return_data z3.Value, is_concrete_condiction bool) bool {
	// Check if there is an entry for current_pc
	if is_concrete_condiction {
		return false
	}

	nextMap, exists := (*visitedNodes)[current_pc]
	if !exists {
		return false
	}

	// Check if there is an entry for next_pc under current_pc
	visitedInfo, exists := nextMap[next_pc]
	if !exists {
		return false
	}

	// method 1： Only compare current_pc, next_pc  => Limitations, important step paths will be lost
	// visitedInfo = visitedInfo
	// return true

	// method 2： compare current_pc, next_pc,stack
	if return_data == nil {
		return_data = SimplifyZ3BV(memory.Context().FromInt(0, memory.Context().IntSort()).(z3.Int).ToBV(256))
	}

	// Iterate through the stacks to find the matching stack
	for i, s := range visitedInfo.Stacks {
		if AreSymbolicStacksEqual(s, stackValue) { //if AreSymbolicStacksEqual(s, stackValue) && AreSymbolicEqual(visitedInfo.Memory[i], memory) {
			// Check if the count for this stackValue is 10 or more
			if visitedInfo.Count[i] >= VISIST_Threshold {
				return true
			}
			break //stop
		}
	}

	return false
}

func updateVisitedNodes(current_pc uint64, next_pc uint64, stackValue *SymbolicStack, visitedNodes *VisitedNodes, memory z3.Value) {
	if _, exists := (*visitedNodes)[current_pc]; !exists { // Ensure the mapping for current_pc exists
		(*visitedNodes)[current_pc] = make(map[uint64]*VisitedInfo)
	}

	if _, exists := (*visitedNodes)[current_pc][next_pc]; !exists { // Ensure the mapping for next_pc exists
		(*visitedNodes)[current_pc][next_pc] = &VisitedInfo{
			Stacks:     make([]*SymbolicStack, 0),
			Count:      make([]int, 0),
			Memory:     make([]z3.Value, 0),
			ReturnData: make([]z3.Value, 0),
		}
	}

	// if return_data == nil {
	// 	return_data = SimplifyZ3BV(memory.Context().FromInt(0, memory.Context().IntSort()).(z3.Int).ToBV(256))
	// }
	node := (*visitedNodes)[current_pc][next_pc]
	stackIndex := -1
	for i, node_stack := range node.Stacks { // Check if stackValue already exists
		if AreSymbolicStacksEqual(node_stack, stackValue) && AreSymbolicEqual(node.Memory[i], memory) {
			stackIndex = i
			break
		}
	}

	if stackIndex == -1 { // New stack state: add to Stacks and set count to 1 in Count
		copy_stack := newstack()
		copy_stack.data = stackValue.Copy().Data()
		node.Stacks = append(node.Stacks, copy_stack)
		node.Count = append(node.Count, 1)
		// node.Memory = append(node.Memory, memory)
		// node.ReturnData = append(node.ReturnData, return_data)
	} else { // Existing stack state: increment the corresponding count
		node.Count[stackIndex]++
	}
}

func DeepCopyScopeContext(original *ScopeContext) *ScopeContext {
	copy := &ScopeContext{}

	if original.Stack != nil {
		copy.Stack = newstack()
		copy.Stack.data = original.Stack.Copy().data
	}
	if original.Contract != nil {
		copy.Contract = original.Contract.Copy() // can be the same
	}
	if original.Memory != nil {
		copy.Memory = NewSymbolicMemory(original.Z3Contex)
		*copy.Memory = original.Memory.Copy()
	}
	if original.Z3Contex != nil {
		copy.Z3Contex = original.Z3Contex // can be the same
	}

	return copy
}

func isEndOfPath(op vm.OpCode) bool {
	switch op {
	case vm.STOP, vm.RETURN, vm.REVERT, vm.INVALID, vm.SELFDESTRUCT:
		return true
	default:
		return false
	}
}

func (in *SymbolicEVMInterpreter) SymbolicRun_DFS(contract *Contract, z3Contex *z3.Context, currentPath ExecutionPath, executionPathList *ExecutionPathList) {
	if len(contract.Code) == 0 {
		ErrNoCode := errors.New("SymbolicRun: there is no code")
		color.Error.Println(ErrNoCode)
		// os.Exit(1)
		return
	}

	if in.evm.depth == 0 {
		in.evm.orgin = contract.caller
	}

	var (
		mem         = NewSymbolicMemory(z3Contex)
		stack       = newstack() // local stack
		callContext = &ScopeContext{
			Memory:   mem,
			Stack:    stack,
			Contract: contract,
			Z3Contex: z3Contex,
		}
		pc = uint64(0) // program counter
	)

	var visitedNodesMap = make(VisitedNodes) //store visited nodes
	copyPath := DeepCopyExecutionPath(currentPath)
	in.SearchPathsByDFS(contract, pc, callContext, &visitedNodesMap, copyPath, executionPathList)
}

func (in *SymbolicEVMInterpreter) SearchPathsByDFS(contract *Contract, pc uint64, callContext *ScopeContext, visitedNodesMap *VisitedNodes, currentPath ExecutionPath, executionPathList *ExecutionPathList) {

	op := contract.GetOp(pc)
	// callContext.Stack.PrintCurrentStack()
	// callContext.Memory.PrintSymbolicMemory()
	// color.Blue.Println("*****Operation******:", op, "*****Code Location*****:", pc, "*****Evm.depth*****:", in.evm.depth)

	var ret z3.Value
	var err error
	switch op {
	case vm.JUMPI:
		current_pc := pc
		next_pc_list := make([]uint64, 0)
		pos, cond := callContext.Stack.pop(), callContext.Stack.pop()
		cond = callContext.Z3Contex.Simplify(cond, cond.Context().Config())
		pos = callContext.Z3Contex.Simplify(pos, pos.Context().Config())

		// record memory,  stack , storage, return(value and err)
		updateExecutionPath(&currentPath, callContext.Stack, callContext.Memory, current_pc, op, ret, err, in.evm.depth, contract.self)
		cond_value, is_concrete_condiction := cond.(z3.BV).AsBigUnsigned()
		var pos_value uint64
		var is_concrete_position bool

		pos = callContext.Z3Contex.Simplify(pos, pos.Context().Config())
		pos_value, _, is_concrete_position = pos.(z3.BV).AsUint64()

		// # Determine the next pc list
		if is_concrete_condiction { // if cond is a concrete value, we just have one next_pc:
			if cond_value.Sign() > 0 { //  cond is  true
				if is_concrete_position && contract.GetOp(pos_value) == vm.JUMPDEST { // only when pos is a concrete value  and next op is JUMPDEST
					next_pc_list = append(next_pc_list, pos_value)
				}

			} else { // cond is  false
				if int(current_pc)+1 < len(contract.Code) {
					next_pc_list = append(next_pc_list, current_pc+1)
				}
			}

		} else { // if cond is not a real value, choose all pathes
			if is_concrete_position && contract.GetOp(pos_value) == vm.JUMPDEST { // only when pos is a concrete value  and next op is JUMPDEST
				next_pc_list = append(next_pc_list, pos_value)
			}
			if int(current_pc)+1 < len(contract.Code) {
				next_pc_list = append(next_pc_list, current_pc+1)
			}

		}

		// color.Error.Println("next_pc_list: ", next_pc_list, "current evm.depth: ", in.evm.depth, current_pc)

		for _, next_pc := range next_pc_list { //note: may have more than 1 successful paths
			if next_pc != current_pc+1 && contract.GetOp(next_pc) != vm.JUMPDEST { //invalid code, then next loop
				color.Error.Println("ErrInvalidJump")
			} else {
				next_CallContext := DeepCopyScopeContext(callContext)
				copyPath := DeepCopyExecutionPath(currentPath)
				if !isVisitedNode(current_pc, next_pc, callContext.Stack, visitedNodesMap, callContext.Memory.Store, in.returnData, is_concrete_condiction) && len(callContext.Stack.data) <= 124 { //Avoiding  loops
					if !is_concrete_condiction {
						updateVisitedNodes(current_pc, next_pc, callContext.Stack, visitedNodesMap, callContext.Memory.Store)
					}

					in.SearchPathsByDFS(contract, next_pc, next_CallContext, visitedNodesMap, copyPath, executionPathList)

				} else {
					color.Info.Println("the node has been visited:", next_pc)
				}
			}

		}

	case vm.CALL, vm.STATICCALL, vm.DELEGATECALL:
		//PauseForSeconds(10000)
		current_pc := pc
		var (
			mem   = NewSymbolicMemory(callContext.Z3Contex)
			stack = newstack() // local stack
		)

		updateExecutionPath(&currentPath, stack, mem, current_pc, op, ret, err, in.evm.depth+1, contract.self) // for CALL opcode,  the ret, err are nil
		var returnDataList CrossContractReturnDataList
		copyPath := DeepCopyExecutionPath(currentPath)
		switch op {
		case vm.CALL:
			returnDataList = opCall(&pc, in, callContext, copyPath, executionPathList) // the call path may have different sub paths
		case vm.STATICCALL:
			returnDataList = opStaticCall(&pc, in, callContext, copyPath, executionPathList)
		case vm.DELEGATECALL:
			returnDataList = opDelegateCall(&pc, in, callContext, copyPath, executionPathList)

		}

		// if len(returnDataList) > 1 {
		// 	color.Println(len(returnDataList))
		// 	for _, returnData := range returnDataList {
		// 		color.Println(returnData.ReturnData)
		// 	}

		// 	PauseForSeconds(20)
		// }

		if len(returnDataList) > 0 { // when  the call is successful

			*executionPathList = (*executionPathList)[:len(*executionPathList)-len(returnDataList)] //  Recovery the executionPathList (because of DFS, The final path is newly added.)
			//current_depth := in.evm.depth
			for _, returnData := range returnDataList { //  traverse all returned paths and continue
				//color.Error.Println("PATH_RETURN")
				path_length := len(returnData.Path)
				err = returnData.Path[path_length-1].Current_return_error
				//in.evm.depth = current_depth - 1
				if err != nil && err.Error() != "return token" && err.Error() != "stop token" { //if it is error, stop (in this code, there is no other error because we only update successful paths)
					color.Error.Println("There is a error")
				} else {

					// Or, continue, and coppy the callContext
					in.returnData = returnData.ReturnData
					returnData.Path[path_length-1].Current_memory = returnData.Scope.Memory.Copy()
					returnData.Path[path_length-1].Current_stack = returnData.Scope.Stack.Copy()
					copyPath := DeepCopyExecutionPath(returnData.Path)
					next_CallContext := DeepCopyScopeContext(returnData.Scope)

					// if len(returnDataList) > 1 {
					// 	color.Println(returnData.ReturnData, "GGGGGGG")
					// 	PauseForSeconds(3)
					// }

					var new_visitedNodesMap = make(VisitedNodes) //important: reset visited nodes
					in.SearchPathsByDFS(contract, current_pc+1, next_CallContext, &new_visitedNodesMap, copyPath, executionPathList)
				}

			}

		}
		return

	default:
		substr := "not defined"
		if !strings.Contains(op.String(), substr) { //if invalid opcode,  stop
			// execute the operation
			operation := in.table[op]
			current_pc := pc
			//color.Error.Println(pc, op)

			ret, err = operation.execute(&pc, in, callContext)
			updateExecutionPath(&currentPath, callContext.Stack, callContext.Memory, current_pc, op, ret, err, in.evm.depth, contract.self) // store memory, stack , pc, op, ret, err

			if int(pc+1) >= len(contract.Code) || isEndOfPath(op) || err != nil { // if it is the end of the block
				// color.Warnln("This path has ended!")
				// color.Warnln("\n\n")
				//if op == vm.RETURN || op == vm.STOP || op == vm.REVERT {  // REVERT will record all paths
				//executionPathList.AddPath(currentPath)}
				if op == vm.RETURN || op == vm.STOP {
					executionPathList.AddPath(currentPath) //  only update successful paths
				}
			} else { // Or, continue, and coppy the callContext
				pc = pc + 1
				next_CallContext := DeepCopyScopeContext(callContext)
				in.SearchPathsByDFS(contract, pc, next_CallContext, visitedNodesMap, currentPath, executionPathList)
			}

		}

	}

}
