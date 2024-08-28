package sevm

import (
	"encoding/hex"
	"errors"

	"math/big"
	"os"
	"strings"

	"github.com/aclements/go-z3/z3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gookit/color"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
	"golang.org/x/crypto/sha3"
)

// / opPush1 is a specialized version of pushN
func opPush1(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	var (
		codeLen = uint64(len(scope.Contract.Code))
		// integer = new(uint256.Int)
	)
	opcpde_bytes := scope.Contract.Code[*pc]
	*pc += 1
	value := int64(0)
	if *pc < codeLen {
		value = int64(scope.Contract.Code[*pc])

	}
	val := scope.Z3Contex.FromBigInt(big.NewInt(value), scope.Z3Contex.BVSort(256)).(z3.BV)

	stack_value := scope.Z3Contex.Simplify(val, scope.Z3Contex.Config())
	scope.Stack.push(stack_value)

	PrintOperationLog(opcpde_bytes, stack_value)
	return nil, nil
}

func opPush0(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	errStopToken := errors.New("invalid opcode")
	return nil, errStopToken
}

func opMstore(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	// pop value of the stack
	offset, value := scope.Stack.pop(), scope.Stack.pop() // offset, data
	//color.Println(offset, value)
	error := scope.Memory.Set32(offset, value)

	if error != nil {
		return nil, error
	}
	opcpde_bytes := scope.Contract.Code[*pc]

	// if *pc == 16782 {

	// 	color.Error.Println(offset, value)
	// 	color.Error.Println("=====opMstore结束后=======")
	// 	scope.Memory.PrintSymbolicMemory()

	// 	PauseForSeconds(1000000)
	// }

	PrintOperationLog(opcpde_bytes, nil)
	return nil, nil
}

func opMstore8(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	// pop value of the stack
	offset, value := scope.Stack.pop(), scope.Stack.pop() // offset, data
	//color.Println(offset, value)
	// get the LSB:
	new_value := SimplifyZ3BV(value.(z3.BV).Extract(int(1*8-1), 0))

	error := scope.Memory.Set(offset, 1, new_value)

	// color.Error.Println(offset, value, new_value)
	// PauseForSeconds(1000000)

	if error != nil {
		return nil, error
	}
	opcpde_bytes := scope.Contract.Code[*pc]

	PrintOperationLog(opcpde_bytes, nil)
	return nil, nil
}

func opCallDataSize(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	bv_size := scope.Contract.Input.Sort().BVSize() // bit size
	byte_size := scope.Z3Contex.FromInt(int64(bv_size/8), scope.Z3Contex.BVSort(256))
	stack_value := scope.Z3Contex.Simplify(byte_size, scope.Z3Contex.Config())
	scope.Stack.push(stack_value)

	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	return nil, nil
}

func opLt(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x, y := scope.Stack.pop(), scope.Stack.pop()

	one_value := scope.Z3Contex.FromInt(1, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	zero_value := scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)

	// x = x.(z3.BV).UToInt()
	// y = y.(z3.BV).UToInt()
	// compare_value := x.(z3.Int).LT(y.(z3.Int)).IfThenElse(one_value, zero_value)
	compare_value := x.(z3.BV).ULT(y.(z3.BV)).IfThenElse(one_value, zero_value)

	stack_value := scope.Z3Contex.Simplify(compare_value, scope.Z3Contex.Config())
	scope.Stack.push(stack_value)
	return nil, nil
}

func opGt(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x, y := scope.Stack.pop(), scope.Stack.pop()

	one_value := scope.Z3Contex.FromInt(1, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	zero_value := scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	// x = x.(z3.BV).UToInt()
	// y = y.(z3.BV).UToInt()
	// compare_value := x.(z3.Int).GT(y.(z3.Int)).IfThenElse(one_value, zero_value)
	compare_value := x.(z3.BV).UGT(y.(z3.BV)).IfThenElse(one_value, zero_value)

	stack_value := scope.Z3Contex.Simplify(compare_value, scope.Z3Contex.Config())
	scope.Stack.push(stack_value)
	return nil, nil
}

func opSGt(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x, y := scope.Stack.pop(), scope.Stack.pop()

	one_value := scope.Z3Contex.FromInt(1, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	zero_value := scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	compare_value := x.(z3.BV).SGT(y.(z3.BV)).IfThenElse(one_value, zero_value)

	stack_value := scope.Z3Contex.Simplify(compare_value, scope.Z3Contex.Config())
	scope.Stack.push(stack_value)
	return nil, nil
}

func opSlt(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	// signed LT
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x, y := scope.Stack.pop(), scope.Stack.pop()

	one_value := scope.Z3Contex.FromInt(1, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	zero_value := scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	compare_value := x.(z3.BV).SLT(y.(z3.BV)).IfThenElse(one_value, zero_value)

	stack_value := scope.Z3Contex.Simplify(compare_value, scope.Z3Contex.Config())
	scope.Stack.push(stack_value)
	return nil, nil
}

func opOr(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	x, y := scope.Stack.pop(), scope.Stack.pop()
	stack_value := SimplifyZ3BV(x.(z3.BV).Or(y.(z3.BV)))
	scope.Stack.push(stack_value)

	return nil, nil
}

func opXor(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	x, y := scope.Stack.pop(), scope.Stack.pop()
	stack_value := SimplifyZ3BV(x.(z3.BV).Xor(y.(z3.BV)))
	scope.Stack.push(stack_value)
	return nil, nil
}

func opByte(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	x, y := scope.Stack.pop(), scope.Stack.pop()
	// get offset
	max := scope.Z3Contex.FromInt(248, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	eight := scope.Z3Contex.FromInt(8, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	offset := max.Sub(x.(z3.BV).Mul(eight))

	//save data
	stack_value := scope.Z3Contex.Simplify(y.(z3.BV).URsh(offset), scope.Z3Contex.Config())
	scope.Stack.push(stack_value)

	return nil, nil
}

func makePush(size uint64, pushByteSize int) executionFunc {

	return func(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

		opcpde_bytes := scope.Contract.Code[*pc]
		codeLen := len(scope.Contract.Code)

		startMin := codeLen
		if int(*pc+1) < startMin {
			startMin = int(*pc + 1)
		}
		endMin := codeLen
		if startMin+pushByteSize < endMin {
			endMin = startMin + pushByteSize
		}
		integer := new(uint256.Int)
		value := integer.SetBytes(common.RightPadBytes(
			scope.Contract.Code[startMin:endMin], pushByteSize))
		stack_value := scope.Z3Contex.FromBigInt(value.ToBig(), scope.Z3Contex.BVSort(256)).(z3.BV)

		scope.Stack.push(stack_value)

		PrintOperationLog(opcpde_bytes, stack_value)
		*pc += size
		return nil, nil
	}
}

func opCallDataLoad(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	offset := scope.Stack.pop()
	offset = scope.Z3Contex.Simplify(offset, offset.Context().Config())

	offset_uint64, _, offset_ok := offset.(z3.BV).AsUint64()
	if !offset_ok {
		errStopToken := errors.New("the offset is not a concrete value")
		color.Error.Println(errStopToken)
		return nil, errStopToken
	}

	scope.Stack.push(getData(scope.Contract.Input, offset_uint64, 32, scope))
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)

	return nil, nil
}

// make swap instruction function
func makeSwap(size int64) executionFunc {
	// switch n + 1 otherwise n would be swapped with n
	size++
	return func(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
		scope.Stack.swap(int(size))
		return nil, nil
	}
}

func opDiv(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x, y := scope.Stack.pop(), scope.Stack.pop()

	stack_value := scope.Z3Contex.Simplify(x.(z3.BV).UDiv(y.(z3.BV)), scope.Z3Contex.Config())
	scope.Stack.push(stack_value)
	return nil, nil
}

func opAdd(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x, y := scope.Stack.pop(), scope.Stack.pop()

	// x = x.(z3.BV).UToInt()
	// y = y.(z3.BV).UToInt()
	// value:= x.(z3.Int).Add(y.(z3.Int)).ToBV(256)
	value := x.(z3.BV).Add(y.(z3.BV))

	stack_value := scope.Z3Contex.Simplify(value, scope.Z3Contex.Config())
	scope.Stack.push(stack_value)
	return nil, nil
}

func opMod(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	x, y := scope.Stack.pop(), scope.Stack.pop()
	stack_value := SimplifyZ3BV(x.(z3.BV).SMod(y.(z3.BV)))
	scope.Stack.push(stack_value)
	return nil, nil
}

func opSub(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x, y := scope.Stack.pop(), scope.Stack.pop()

	// x = x.(z3.BV).UToInt()
	// y = y.(z3.BV).UToInt()
	// value:= x.(z3.Int).Sub(y.(z3.Int)).ToBV(256)
	value := x.(z3.BV).Sub(y.(z3.BV))

	stack_value := scope.Z3Contex.Simplify(value, scope.Z3Contex.Config())
	scope.Stack.push(stack_value)
	return nil, nil
}

func opMul(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x, y := scope.Stack.pop(), scope.Stack.pop() // x or y can be bool
	x = scope.Z3Contex.Simplify(x, x.Context().Config())
	y = scope.Z3Contex.Simplify(y, y.Context().Config())

	// x = x.(z3.BV).UToInt()
	// y = y.(z3.BV).UToInt()
	// value:= x.(z3.Int).Mul(y.(z3.Int)).ToBV(256)
	value := x.(z3.BV).Mul(y.(z3.BV))

	stack_value := scope.Z3Contex.Simplify(value, scope.Z3Contex.Config())

	scope.Stack.push(stack_value)
	return nil, nil

}

func opExp(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x, y := scope.Stack.pop(), scope.Stack.pop()

	//
	x_value, _, x_ok := x.(z3.BV).AsUint64()
	y_value, _, y_ok := y.(z3.BV).AsUint64()
	if x_ok && y_ok {
		x_big := big.NewInt(int64(x_value))
		y_big := big.NewInt(int64(y_value))
		result_big := new(big.Int).Exp(x_big, y_big, nil)
		value := scope.Z3Contex.FromBigInt(result_big, scope.Z3Contex.BVSort(256)).(z3.BV)

		scope.Stack.push(SimplifyZ3BV(value))

	} else { // if x , y are not concrete values, then use symbols

		x = x.(z3.BV).UToInt()
		y = y.(z3.BV).UToInt()
		scope.Stack.push(SimplifyZ3BV(x.(z3.Int).Exp(y.(z3.Int)).ToBV(256)))
	}

	return nil, nil
}

func opAnd(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x, y := scope.Stack.pop(), scope.Stack.pop()

	stack_value := scope.Z3Contex.Simplify(x.(z3.BV).And(y.(z3.BV)), scope.Z3Contex.Config())

	scope.Stack.push(stack_value)
	return nil, nil
}

func makeDup(size int64) executionFunc {
	return func(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
		opcpde_bytes := scope.Contract.Code[*pc]
		PrintOperationLog(opcpde_bytes, nil)

		if int(size) <= scope.Stack.Len() {
			scope.Stack.dup(int(size))
			return nil, nil
		} else {
			err := errors.New("panic: runtime error: index out of range")
			color.Error.Println(err)
			return nil, err

		}

	}
}

func opSHL(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := scope.Stack.pop(), scope.Stack.pop()

	stack_value := scope.Z3Contex.Simplify(value.(z3.BV).Lsh(shift.(z3.BV)), scope.Z3Contex.Config())
	scope.Stack.push(stack_value)
	return nil, nil
}

func opEq(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x, y := scope.Stack.pop(), scope.Stack.pop()

	one_value := scope.Z3Contex.FromInt(1, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	zero_value := scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	compare_value := x.(z3.BV).Eq(y.(z3.BV)).IfThenElse(one_value, zero_value)

	stack_value := scope.Z3Contex.Simplify(compare_value, scope.Z3Contex.Config())

	scope.Stack.push(stack_value)
	return nil, nil
}

func opNot(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x := scope.Stack.pop()
	stack_value := scope.Z3Contex.Simplify(x.(z3.BV).Not(), scope.Z3Contex.Config())
	scope.Stack.push(stack_value)
	return nil, nil
}

func opJumpdest(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	return nil, nil
}

func opJump(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	pos := scope.Stack.pop().(z3.BV)
	//pos := scope.Stack.pop().(z3.BV).UToInt()
	pos_value, _, ok := scope.Z3Contex.Simplify(pos, pos.Context().Config()).(z3.BV).AsUint64()

	if !ok {
		err := errors.New("the code position is not a concrete value")
		color.Error.Println(err)
		return nil, err
	}

	if !scope.Contract.validJumpdest(pos_value) {
		return nil, vm.ErrInvalidJump
	}
	*pc = pos_value - 1 // pc will be increased by the interpreter loop

	return nil, nil
}

func opStop(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	errStopToken := errors.New("stop token")
	return nil, errStopToken

}

func opCallValue(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)

	scope.Stack.push(SimplifyZ3BV(scope.Contract.value))
	return nil, nil
}

func opIszero(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	x := SimplifyZ3BV(scope.Stack.pop().(z3.BV))
	one_value := scope.Z3Contex.FromInt(1, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
	zero_value := scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)

	//zero_int_value := scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort())
	//compare_value := x.(z3.Int).Eq(zero_int_value.(z3.Int)).IfThenElse(one_value, zero_value)
	compare_value := x.Eq(zero_value).IfThenElse(one_value, zero_value)

	scope.Stack.push(SimplifyZ3BV(compare_value.(z3.BV)))

	return nil, nil
}

func opRevert(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)

	offset, size := scope.Stack.pop(), scope.Stack.pop()
	size_int64, _, ok := size.(z3.BV).AsUint64()
	if !ok {
		errStopToken := errors.New("inSize is not a concrete value")
		color.Error.Println("inSize is not a concrete value")
		return nil, errStopToken

	}

	offset_uint64, _, ok := offset.(z3.BV).AsUint64()
	if !ok { // can't support Z3.Value
		errStopToken := errors.New("offset is not a concrete value")
		color.Error.Println("offset is not a concrete value", offset_uint64)
		return nil, errStopToken
	}

	ret := scope.Memory.GetCopy(offset, size_int64) // size can be different
	interpreter.returnData = ret

	return ret, vm.ErrExecutionReverted

}

func opExtCodeSize(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	slot := scope.Stack.pop()

	_, data_ok := slot.(z3.BV).AsBigUnsigned()
	if !data_ok {
		variable_name := removeBackslashesAndPipes("(EXTCODESIZE " + slot.String() + ")")
		var loc_data z3.BV
		if _, exists := interpreter.Z3Variables[variable_name]; exists {
			loc_data = interpreter.Z3Variables[variable_name]
		} else {
			loc_data = scope.Z3Contex.BVConst(variable_name, 256)
			interpreter.Z3Variables[variable_name] = loc_data
		}
		scope.Stack.push(loc_data)

	} else {
		addressHash_string := "0x" + slot.String()[26:66] // get  the address of contract
		code := GetContractCodeByAddress(addressHash_string)
		length := len(code)
		length_value := scope.Z3Contex.FromInt(int64(length), scope.Z3Contex.BVSort(256))
		scope.Stack.push(length_value)
	}

	return nil, nil
}

func opMload(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	// pop value of the stack

	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	offset := scope.Stack.pop()

	offset_uint64, _, ok := offset.(z3.BV).AsUint64()
	if !ok { // can't support Z3.Value
		errStopToken := errors.New("offset is not a concrete value")
		color.Error.Println("offset is not a concrete value", offset_uint64)
		return nil, errStopToken
	}

	m_data := scope.Memory.GetCopy(offset, 32)
	if m_data == nil {
		errStopToken := errors.New("m_data is nil")
		return nil, errStopToken

	}

	// length := m_data.Sort().BVSize()
	// if length < 256 {
	// 	space := scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256 - length) // a9059cbb
	// 	m_data = space.Concat(m_data.(z3.BV))
	// }

	stack_value := scope.Z3Contex.Simplify(m_data, scope.Z3Contex.Config())
	scope.Stack.push(stack_value)

	return nil, nil

}

func opPop(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	scope.Stack.pop()
	return nil, nil
}

func opCaller(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	caller := scope.Contract.caller
	scope.Stack.push(caller)

	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	return nil, nil
}

func opKeccak256(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	// scope.Stack.PrintCurrentStack()
	// scope.Memory.PrintSymbolicMemory()
	offset, size := scope.Stack.pop(), scope.Stack.pop()

	size_int64, _, ok := size.(z3.BV).AsUint64()
	if !ok {
		errStopToken := errors.New("inSize is not a concrete value")
		color.Error.Println(errStopToken)
		return nil, errStopToken

	}
	data := SimplifyZ3BV(scope.Memory.GetCopy(offset, size_int64).(z3.BV))

	_, data_ok := data.AsBigUnsigned()
	if !data_ok {
		variable_name := removeBackslashesAndPipes("SHA3[" + data.String() + "]")
		var loc_data z3.BV
		if _, exists := interpreter.Z3Variables[variable_name]; exists {
			loc_data = interpreter.Z3Variables[variable_name]
		} else {
			loc_data = scope.Z3Contex.BVConst(variable_name, 256)
			interpreter.Z3Variables[variable_name] = loc_data
		}
		scope.Stack.push(loc_data)

	} else {
		data_string := strings.TrimPrefix(data.String(), "#x")
		data, err := hex.DecodeString(data_string)
		if err != nil {
			panic(err)
		}
		hasher := sha3.NewLegacyKeccak256()
		hasher.Write(data)
		hash := hasher.Sum(nil)
		hashHex := hex.EncodeToString(hash)
		stack_value := scope.Z3Contex.FromBigInt(common.HexToHash(hashHex).Big(), scope.Z3Contex.BVSort(256)).(z3.BV)
		scope.Stack.push(stack_value)

	}

	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)

	return nil, nil
}

func opSload(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)
	key := scope.Stack.pop()
	_, key_ok := key.(z3.BV).AsBigUnsigned()

	var loc_data z3.BV
	if !key_ok { //if load can not  be loaded (related to intput)

		variable_name := removeBackslashesAndPipes("SLOAD" + scope.Contract.self.String() + "=>" + key.String())

		// color.Info.Println(variable_name)
		// color.Info.Println(scope.Z3Contex.BVConst(variable_name, 256))
		if _, exists := interpreter.Z3Variables[variable_name]; exists {
			loc_data = interpreter.Z3Variables[variable_name]
		} else {
			loc_data = scope.Z3Contex.BVConst(variable_name, 256)
			interpreter.Z3Variables[variable_name] = loc_data
		}

	} else { //if load can be loaded (Unrelated to intput)

		block_hash := common.HexToHash("0xc4a6979d83031170518df33017fa64690546bc07696c3b81a8f02e795e2cbea6") //18743597
		//block_hash := common.HexToHash("0x89e9e608b092afb1257b74fb071ae237c2e90769c0ab2904c64aa8b111501568") //14684307

		contract_string := "0x" + scope.Contract.self.String()[26:66]
		contract_hash := common.HexToAddress(contract_string)

		key_string := "0x" + key.String()[2:66]
		key_hash := common.HexToHash(key_string)

		key_data := GetStorageAtHash(contract_hash, key_hash, block_hash)

		loc_data = scope.Z3Contex.FromBigInt(new(big.Int).SetBytes(key_data), scope.Z3Contex.BVSort(256)).(z3.BV)

		if !IsSmartContractAddress("0x" + hex.EncodeToString(key_data)) { // if it is not smart contract
			variable_name := removeBackslashesAndPipes("SLOAD" + scope.Contract.self.String() + "=>" + key.String())
			if _, exists := interpreter.Z3Variables[variable_name]; exists {
				loc_data = interpreter.Z3Variables[variable_name]
			} else {
				loc_data = scope.Z3Contex.BVConst(variable_name, 256)
				interpreter.Z3Variables[variable_name] = loc_data
			}
		}

	}

	scope.Stack.push(SimplifyZ3BV(loc_data))

	return nil, nil
}

func opSstore(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	scope.Stack.pop()
	scope.Stack.pop()

	// if interpreter.evm.depth > 0 {
	// 	color.Info.Println("============================================================")
	// 	color.Info.Println(value)
	// }

	//location, value := scope.Stack.pop(), scope.Stack.pop()
	// color.Error.Println(location)
	// color.Error.Println(value)
	// if interpreter.evm.depth > 0 {
	// 	PauseForSeconds(100)
	// }

	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)

	return nil, nil
}

func opSHR(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	// Note, second operand is left in the stack; accumulate result into it, and no need to push it afterwards
	shift, value := scope.Stack.pop(), scope.Stack.pop()
	stack_value := scope.Z3Contex.Simplify(value.(z3.BV).URsh(shift.(z3.BV)), scope.Z3Contex.Config())
	scope.Stack.push(stack_value)

	return nil, nil

}

func makeLog(size int) executionFunc {
	return func(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
		opcpde_bytes := scope.Contract.Code[*pc]
		PrintOperationLog(opcpde_bytes, nil)

		stack := scope.Stack
		_, _ = stack.pop(), stack.pop()
		for i := 0; i < size; i++ {
			stack.pop()

		}

		// _ = scope.Memory.GetCopy(mStart, mSize)

		return nil, nil
	}
}

func opReturn(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	offset, size := scope.Stack.pop(), scope.Stack.pop()
	size_int64, _, ok := size.(z3.BV).AsUint64()
	if !ok {
		errStopToken := errors.New("inSize is not a concrete value")
		color.Error.Println(errStopToken)
		return nil, errStopToken

	}

	ret := scope.Memory.GetCopy(offset, size_int64) // size can be different
	errStopToken := errors.New("return token")
	return ret, errStopToken
}

type CrossContractReturnData struct {
	Scope      *ScopeContext
	Path       ExecutionPath
	ReturnData z3.Value
}
type CrossContractReturnDataList []CrossContractReturnData

func opCall(pc *uint64, interpreter *SymbolicEVMInterpreter, orginal_scope *ScopeContext, currentPath ExecutionPath, executionPathList *ExecutionPathList) CrossContractReturnDataList {

	scope := DeepCopyScopeContext(orginal_scope)
	stack := scope.Stack

	stack.pop() //GAS
	// Pop other call parameters.
	toAddr, value, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	// Get the arguments from the memory.
	inSize_int64, _, ok := inSize.(z3.BV).AsUint64()
	if !ok {
		errStopToken := errors.New("inSize is not a concrete value")
		color.Error.Println(errStopToken)
		os.Exit(1)

	}
	args := scope.Memory.GetCopy(inOffset, inSize_int64) //Input of Called contracts

	// PauseForSeconds(1000)
	// color.Info.Println("Call", inOffset, inSize)
	// color.Info.Println("Call", inOffset, inSize_int64, args)
	// to_string := SimplifyZ3BV(toAddr.(z3.BV)).String()
	// color.Println(to_string)
	original_list_length := len(*executionPathList)

	var returnDataList = make(CrossContractReturnDataList, 0)
	if args == nil { // When there is an ETH transfer, args  can be nil
		path_scope := DeepCopyScopeContext(scope)
		color.Println(original_list_length)
		executionPathList.AddPath(currentPath)
		current_path := DeepCopyExecutionPath(currentPath)
		returndata := CrossContractReturnData{
			Scope:      path_scope,
			Path:       current_path,
			ReturnData: nil,
		}

		returnDataList = append(returnDataList, returndata)
		return returnDataList
	}
	interpreter.evm.Call(scope.Contract.self, toAddr.(z3.BV), args.(z3.BV), value.(z3.BV), currentPath, executionPathList)

	current_list_length := len(*executionPathList)

	one_value := SimplifyZ3BV(scope.Z3Contex.FromInt(1, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256))
	zero_value := SimplifyZ3BV(scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256))

	if current_list_length-original_list_length > 0 {
		for path_index := 0; path_index < (current_list_length - original_list_length); path_index++ {
			path_scope := DeepCopyScopeContext(scope)
			current_path := DeepCopyExecutionPath((*executionPathList)[original_list_length+path_index])
			path_length := len(current_path)
			err := current_path[path_length-1].Current_return_error
			ret := current_path[path_length-1].Current_return_value

			var temp z3.Value
			if err != nil && err.Error() != "return token" && err.Error() != "stop token" {
				temp = zero_value
			} else {
				temp = one_value
			}

			path_scope.Stack.push(temp)
			// if err == nil || err == ErrExecutionReverted {
			retSize_uint64, _, ok := retSize.(z3.BV).AsUint64()
			if !ok {
				errStopToken := errors.New("retSize is not a concrete value")
				color.Error.Println(errStopToken)
				os.Exit(1)

			}

			if err == nil || err == vm.ErrExecutionReverted || err.Error() == "return token" || err.Error() == "stop token" {
				path_scope.Memory.Set(retOffset, retSize_uint64, ret)
			}

			returndata := CrossContractReturnData{
				Scope:      path_scope,
				Path:       current_path,
				ReturnData: ret,
			}

			returnDataList = append(returnDataList, returndata)

		}

	}

	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)

	return returnDataList

}

func opStaticCall(pc *uint64, interpreter *SymbolicEVMInterpreter, orginal_scope *ScopeContext, currentPath ExecutionPath, executionPathList *ExecutionPathList) CrossContractReturnDataList {
	// Pop gas. The actual gas is in interpreter.evm.callGasTemp.

	scope := DeepCopyScopeContext(orginal_scope)
	stack := scope.Stack
	stack.pop()
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := addr

	inSize_int64, _, ok := inSize.(z3.BV).AsUint64()
	if !ok {
		errStopToken := errors.New("inSize is not a concrete value")
		color.Error.Println(errStopToken)
		os.Exit(1)

	}
	args := scope.Memory.GetCopy(inOffset, inSize_int64) //Input of Called contracts
	//ret, returnGas, err := interpreter.evm.StaticCall(scope.Contract, toAddr, args, gas)
	original_list_length := len(*executionPathList)
	interpreter.evm.StaticCall(scope.Contract.self, toAddr.(z3.BV), args.(z3.BV), currentPath, executionPathList) /// return value of the call
	current_list_length := len(*executionPathList)

	one_value := SimplifyZ3BV(scope.Z3Contex.FromInt(1, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256))
	zero_value := SimplifyZ3BV(scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256))

	var returnDataList = make(CrossContractReturnDataList, 0)
	if current_list_length-original_list_length > 0 {
		for path_index := 0; path_index < (current_list_length - original_list_length); path_index++ {
			path_scope := DeepCopyScopeContext(scope)
			current_path := (*executionPathList)[original_list_length+path_index]
			path_length := len(current_path)
			err := current_path[path_length-1].Current_return_error
			ret := current_path[path_length-1].Current_return_value

			var temp z3.Value
			if err != nil && err.Error() != "return token" && err.Error() != "stop token" {
				temp = zero_value
			} else {
				temp = one_value
			}

			path_scope.Stack.push(SimplifyZ3BV(temp.(z3.BV)))
			// if err == nil || err == ErrExecutionReverted {
			retSize_uint64, _, ok := retSize.(z3.BV).AsUint64()
			if !ok {
				errStopToken := errors.New("retSize is not a concrete value")
				color.Error.Println(errStopToken)
				os.Exit(1)

			}

			if err == nil || err == vm.ErrExecutionReverted || err.Error() == "return token" || err.Error() == "stop token" {

				path_scope.Memory.Set(retOffset, retSize_uint64, ret)

			}

			returndata := CrossContractReturnData{
				Scope:      path_scope,
				Path:       current_path,
				ReturnData: ret,
			}
			returnDataList = append(returnDataList, returndata)
		}
	}

	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)

	return returnDataList

}

func opDelegateCall(pc *uint64, interpreter *SymbolicEVMInterpreter, orginal_scope *ScopeContext, currentPath ExecutionPath, executionPathList *ExecutionPathList) CrossContractReturnDataList {

	scope := DeepCopyScopeContext(orginal_scope)
	stack := scope.Stack
	stack.pop()
	// Pop other call parameters.
	addr, inOffset, inSize, retOffset, retSize := stack.pop(), stack.pop(), stack.pop(), stack.pop(), stack.pop()
	toAddr := addr
	// Get arguments from the memory.
	inSize_int64, _, ok := inSize.(z3.BV).AsUint64()
	if !ok {
		errStopToken := errors.New("inSize is not a concrete value")
		color.Error.Println(errStopToken)
		os.Exit(1)

	}
	args := scope.Memory.GetCopy(inOffset, inSize_int64) //Input of Called contracts

	original_list_length := len(*executionPathList)

	// self is the caller
	interpreter.evm.DelegateCall(scope.Contract.caller, scope.Contract.self, toAddr.(z3.BV), args.(z3.BV), currentPath, executionPathList) /// retuen value of the call
	current_list_length := len(*executionPathList)

	var returnDataList = make(CrossContractReturnDataList, 0)
	if current_list_length-original_list_length > 0 {
		for path_index := 0; path_index < (current_list_length - original_list_length); path_index++ {
			path_scope := DeepCopyScopeContext(scope)
			current_path := (*executionPathList)[original_list_length+path_index]
			path_length := len(current_path)
			err := current_path[path_length-1].Current_return_error
			ret := current_path[path_length-1].Current_return_value

			one_value := scope.Z3Contex.FromInt(1, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
			zero_value := scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort()).(z3.Int).ToBV(256)
			var temp z3.Value
			if err != nil && err.Error() != "return token" || err.Error() != "stop token" {
				temp = zero_value
			} else {
				temp = one_value
			}

			path_scope.Stack.push(temp)
			// if err == nil || err == ErrExecutionReverted {
			retSize_uint64, _, ok := retSize.(z3.BV).AsUint64()
			if !ok {
				errStopToken := errors.New("retSize is not a concrete value")
				color.Error.Println(errStopToken)
				os.Exit(1)

			}

			if err == nil || err == vm.ErrExecutionReverted || err.Error() == "return token" && err.Error() == "stop token" {
				path_scope.Memory.Set(retOffset, retSize_uint64, ret)
			}

			returndata := CrossContractReturnData{
				Scope:      path_scope,
				Path:       current_path,
				ReturnData: ret,
			}
			returnDataList = append(returnDataList, returndata)

		}
	}

	opcpde_bytes := scope.Contract.Code[*pc]
	PrintOperationLog(opcpde_bytes, nil)

	return returnDataList

}

func opTimestamp(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	variable_name := "Block_Time"
	var loc_data z3.BV
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		loc_data = interpreter.Z3Variables[variable_name]
	} else {
		loc_data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = loc_data
	}

	scope.Stack.push(loc_data)

	return nil, nil
}

func getData(inputdata z3.BV, start uint64, size uint64, scope *ScopeContext) z3.BV {
	length := uint64(inputdata.Sort().BVSize())
	start_pos := start * 8
	end_pos := start*8 + size*8

	if int(length)-int(start_pos)-1 <= 0 {
		zero := scope.Z3Contex.FromInt(0, scope.Z3Contex.IntSort()).(z3.Int).ToBV(int(size * 8))
		return SimplifyZ3BV(zero)

	}
	if int(length)-int(end_pos) < 0 {
		data := inputdata.Extract(int(length-start_pos-1), 0)
		lacked_bv := data.Context().FromInt(0, data.Context().BVSort(int(size*8)-data.Sort().BVSize())).(z3.BV)
		data = data.Concat(lacked_bv)
		return SimplifyZ3BV(data)

	} else {
		data := inputdata.Extract(int(length-start_pos-1), int(length-end_pos))
		return SimplifyZ3BV(data)
	}

}

func ReverseZ3BV(orginal_bv z3.BV) z3.BV {
	length := orginal_bv.Sort().BVSize()
	start_bv := orginal_bv.Context().FromInt(0, orginal_bv.Context().BVSort(4)).(z3.BV)
	for offset := 0; offset < length; offset += 256 {
		if offset+256 < length {
			value := orginal_bv.Extract(offset+256-1, offset)
			start_bv = start_bv.Concat(value)
		} else {
			value := orginal_bv.Extract(length-1, offset)
			start_bv = start_bv.Concat(value)
		}

	}
	length = start_bv.Sort().BVSize()
	start_bv = start_bv.Extract(length-1-4, 0)

	return SimplifyZ3BV(start_bv)

}

func opCallDataCopy(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	var (
		memOffset  = scope.Stack.pop()
		dataOffset = scope.Stack.pop()
		length     = scope.Stack.pop()
	)
	dataOffset_uint64, _, ok := dataOffset.(z3.BV).AsUint64()
	if !ok {
		color.Error.Println("dataOffset is not a concrete value!")
		errStopToken := errors.New("dataOffset is not a concrete value")
		return nil, errStopToken

	}
	length_uint64, _, ok := length.(z3.BV).AsUint64()
	if !ok {
		color.Error.Println("length is not a concrete value!")
		errStopToken := errors.New("length is not a concrete value")
		return nil, errStopToken

	}
	//color.Error.Println(memOffset, length_uint64, getData(scope.Contract.Input, dataOffset_uint64, length_uint64, scope))
	scope.Memory.Set(memOffset, length_uint64, getData(scope.Contract.Input, dataOffset_uint64, length_uint64, scope))

	// if *pc == 8870 {
	// 	color.Error.Println("Input: ", scope.Contract.Input)
	// 	color.Error.Println("POS: ", dataOffset_uint64, length_uint64)
	// 	color.Error.Println("Stack data:", dataOffset_uint64, length_uint64)
	// 	color.Error.Println("Idal memory data:", "000000000000000000000000c02aaa39b223fe8d0a0e5c4f27ead9083c756cc20000000000000000000000004306b12f8e824ce1fa9604bbd88f2ad4f0fe3c54")
	// 	color.Error.Println("Now data:", getData(scope.Contract.Input, dataOffset_uint64, length_uint64, scope))
	// 	PauseForSeconds(10000000000000)
	// }

	return nil, nil
}

func opGas(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	var loc_data z3.BV
	variable_name := "GAS"
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		loc_data = interpreter.Z3Variables[variable_name]
	} else {
		loc_data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = loc_data
	}
	scope.Stack.push(loc_data)

	return nil, nil
}

func opReturnDataSize(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	length := 0
	if interpreter.returnData != nil {
		length = interpreter.returnData.Sort().BVSize() / 8
	}
	length_BV := scope.Z3Contex.FromInt(int64(length), scope.Z3Contex.BVSort(256))

	//return_data_string := interpreter.returnData.String()
	//color.Println(return_data_string)

	scope.Stack.push(length_BV)
	return nil, nil
}

func opReturnDataCopy(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	var (
		memOffset  = scope.Stack.pop()
		dataOffset = scope.Stack.pop()
		length     = scope.Stack.pop()
	)

	offset64_uint64, _, ok := dataOffset.(z3.BV).AsUint64()
	if !ok {
		color.Error.Println("dataOffset is not a concrete value!")
		errStopToken := errors.New("length is not a concrete value")
		return nil, errStopToken

	}

	length_uint64, _, ok := length.(z3.BV).AsUint64()
	if !ok {
		color.Error.Println("length is not a concrete value!")
		errStopToken := errors.New("length is not a concrete value")
		return nil, errStopToken

	}

	var end = offset64_uint64 + length_uint64

	if interpreter.returnData == nil {
		interpreter.returnData = scope.Z3Contex.FromInt(0, scope.Z3Contex.BVSort(256))
	}

	var returndata z3.BV

	if int(end-1) <= 0 || (int(end-1) < int(offset64_uint64)) {
		returndata = scope.Z3Contex.FromInt(0, scope.Z3Contex.BVSort(256)).(z3.BV)

	} else {

		if end > uint64(interpreter.returnData.Sort().BVSize()) || offset64_uint64 > uint64(interpreter.returnData.Sort().BVSize()) {
			color.Error.Println("length is not a right value")
			errStopToken := errors.New("length is not a right value")
			return nil, errStopToken
		} else {
			//color.Println(int(end-1), int(offset64_uint64), interpreter.returnData)
			returndata = interpreter.returnData.(z3.BV).Extract(int(end-1), int(offset64_uint64))

		}

	}

	if length_uint64 > 0 {
		scope.Memory.Set(memOffset, length_uint64, returndata)

	}

	return nil, nil

}

func getBytesData(data []byte, start uint64, size uint64) []byte {
	length := uint64(len(data))
	if start > length {
		start = length
	}
	end := start + size
	if end > length {
		end = length
	}
	return common.RightPadBytes(data[start:end], int(size))
}

func opCodeCopy(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	var (
		memOffset  = scope.Stack.pop()
		codeOffset = scope.Stack.pop()
		length     = scope.Stack.pop()
	)

	codeOffset_uint64, _, ok := codeOffset.(z3.BV).AsUint64()
	if !ok {
		color.Error.Println("codeOffset is not a concrete value!")
		errStopToken := errors.New("codeOffset is not a concrete value")
		return nil, errStopToken

	}
	length_uint64, _, ok := length.(z3.BV).AsUint64()
	if !ok {
		color.Error.Println("length is not a concrete value!")
		errStopToken := errors.New("length is not a concrete value")
		return nil, errStopToken

	}

	codeCopy := getBytesData(scope.Contract.Code, codeOffset_uint64, length_uint64)
	hexString := hex.EncodeToString(codeCopy)
	bv_data := scope.Z3Contex.FromBigInt(common.HexToHash(hexString).Big(), scope.Z3Contex.BVSort(int(length_uint64)*8)).(z3.BV)
	scope.Memory.Set(memOffset, length_uint64, bv_data)

	return nil, nil
}

func opAddress(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	scope.Stack.push(scope.Contract.self)
	return nil, nil
}

func opInvalid(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	return nil, nil
}

func opGasLimit(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	var loc_data z3.BV
	variable_name := "GAS_LIMITED"
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		loc_data = interpreter.Z3Variables[variable_name]
	} else {
		loc_data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = loc_data
	}
	scope.Stack.push(loc_data)
	return nil, nil
}

func opGasPrice(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	var loc_data z3.BV
	variable_name := "GAS_PRICE"
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		loc_data = interpreter.Z3Variables[variable_name]
	} else {
		loc_data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = loc_data
	}
	scope.Stack.push(loc_data)
	return nil, nil
}

func opNumber(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	var loc_data z3.BV
	variable_name := "BLOCK_NUMBER"
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		loc_data = interpreter.Z3Variables[variable_name]
	} else {
		loc_data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = loc_data
	}
	scope.Stack.push(loc_data)
	return nil, nil
}

func opSelfBalance(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	variable_name := "(ETHBalance [" + scope.Contract.self.String() + "])"
	var data z3.BV
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		data = interpreter.Z3Variables[variable_name]
	} else {
		data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = data
	}
	scope.Stack.push(data)
	return nil, nil

}

func opBalance(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	address := scope.Stack.pop()
	variable_name := removeBackslashesAndPipes("(ETHBalance [" + address.String() + "])")
	var data z3.BV
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		data = interpreter.Z3Variables[variable_name]
	} else {
		data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = data
	}
	scope.Stack.push(data)
	return nil, nil

}

func opChainID(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	variable_name := "ChainID"
	var data z3.BV
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		data = interpreter.Z3Variables[variable_name]
	} else {
		data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = data
	}
	scope.Stack.push(data)
	return nil, nil
}

func opOrigin(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	orgin := interpreter.evm.orgin
	scope.Stack.push(orgin)
	return nil, nil

}

func opExtCodeHash(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	key := scope.Stack.pop()

	variable_name := removeBackslashesAndPipes("(ExtCodeHash [" + key.String() + "])")
	var data z3.BV
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		data = interpreter.Z3Variables[variable_name]
	} else {
		data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = data
	}
	scope.Stack.push(data)

	return nil, nil
}

func opSignExtend(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {
	back, num := scope.Stack.pop(), scope.Stack.pop()

	num_int, _, ok := num.(z3.BV).AsUint64()
	if !ok {
		errStopToken := errors.New("num is not a concrete value")
		color.Error.Println(errStopToken)
		return nil, errStopToken

	}
	stack_value := SimplifyZ3BV(back.(z3.BV).SignExtend(int(num_int)))
	scope.Stack.push(stack_value)

	return nil, nil
}

func opCoinbase(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	variable_name := "CoinBase"
	var data z3.BV
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		data = interpreter.Z3Variables[variable_name]
	} else {
		data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = data
	}
	scope.Stack.push(data)
	return nil, nil
}

func opDifficulty(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	variable_name := "Difficulty"
	var data z3.BV
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		data = interpreter.Z3Variables[variable_name]
	} else {
		data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = data
	}
	scope.Stack.push(data)
	return nil, nil
}

func opBlockHash(pc *uint64, interpreter *SymbolicEVMInterpreter, scope *ScopeContext) (z3.Value, error) {

	number := scope.Stack.pop()
	variable_name := removeBackslashesAndPipes("(BLOCKHASH " + number.String() + ")")
	var data z3.BV
	if _, exists := interpreter.Z3Variables[variable_name]; exists {
		data = interpreter.Z3Variables[variable_name]
	} else {
		data = scope.Z3Contex.BVConst(variable_name, 256)
		interpreter.Z3Variables[variable_name] = data
	}
	scope.Stack.push(data)
	return nil, nil
}
