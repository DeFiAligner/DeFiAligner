package sevm

import (
	"errors"
	"github.com/aclements/go-z3/z3"
	"github.com/gookit/color"
)

type SEVM struct {
	// Context provides auxiliary blockchain related information

	// used throughout the execution of the tx.
	interpreter *SymbolicEVMInterpreter
	orgin       z3.BV

	depth int

	// abort is used to abort the EVM calling operations

}

func (evm *SEVM) Call(caller z3.BV, addr z3.BV, input z3.BV, value z3.BV, currentPath ExecutionPath, executionPathList *ExecutionPathList) {

	// //update code
	addressHash_string := "0x" + addr.String()[26:66] // get  the address of contract
	code := GetContractCodeByAddress(addressHash_string)
	//color.Error.Println(caller)
	if len(code) == 0 {
		ErrNoCode := errors.New("Call: there is no code")
		color.Error.Println(ErrNoCode)
	} else {
		// If the account has no code, we can abort here
		// NewContract(caller z3.Value, object ContractRef, value *big.Int, gas uint64) *Contract
		contract := NewContract(caller, addr, value, 0)
		contract.SetCallCode(addr, code)
		contract.Input = input
		evm.depth = evm.depth + 1
		evm.interpreter.SymbolicRun_DFS(contract, caller.Context(), currentPath, executionPathList)
		evm.depth = evm.depth - 1
	}

}

func (evm *SEVM) StaticCall(caller z3.BV, addr z3.BV, input z3.BV, currentPath ExecutionPath, executionPathList *ExecutionPathList) {

	// //update code
	addressHash_string := "0x" + addr.String()[26:66] // get  the address of contract
	code := GetContractCodeByAddress(addressHash_string)

	// color.Println(code)
	if len(code) == 0 {
		ErrNoCode := errors.New("StaticCall: there is no code")
		color.Error.Println(ErrNoCode)
	} else {
		// If the account has no code, we can abort here
		// NewContract(caller z3.Value, object ContractRef, value *big.Int, gas uint64) *Contract
		zero_value := input.Context().FromInt(0, input.Context().IntSort()).(z3.Int).ToBV(256)
		contract := NewContract(caller, addr, zero_value, 0)
		contract.Input = input
		contract.SetCallCode(addr, code)

		evm.depth = evm.depth + 1
		evm.interpreter.SymbolicRun_DFS(contract, caller.Context(), currentPath, executionPathList)
		evm.depth = evm.depth - 1

	}

}

func (evm *SEVM) DelegateCall(caller z3.BV, self_address z3.BV, delegateCall_addr z3.BV, input z3.BV, currentPath ExecutionPath, executionPathList *ExecutionPathList) {
	// //update code
	addressHash_string := "0x" + delegateCall_addr.String()[26:66] // get  the address of contract
	code := GetContractCodeByAddress(addressHash_string)
	if len(code) == 0 {
		ErrNoCode := errors.New("DelegateCall: there is no code")
		color.Error.Println(ErrNoCode)
	} else {
		// If the account has no code, we can abort here
		// NewContract(caller z3.Value, object ContractRef, value *big.Int, gas uint64) *Contract
		zero_value := input.Context().FromInt(0, input.Context().IntSort()).(z3.Int).ToBV(256)
		contract := NewContract(caller, self_address, zero_value, 0)
		contract.Input = input
		contract.SetCallCode(delegateCall_addr, code)

		evm.depth = evm.depth + 1
		evm.interpreter.SymbolicRun_DFS(contract, caller.Context(), currentPath, executionPathList)
		evm.depth = evm.depth - 1
	}
}
