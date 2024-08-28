package sevm

import (
	"github.com/aclements/go-z3/z3"
	"github.com/ethereum/go-ethereum/core/vm"
)

type executionFunc func(
	pc *uint64,
	interpreter *SymbolicEVMInterpreter,
	callContext *ScopeContext,
) (
	z3.Value,
	error,
)

//type memorySizeFunc func(*SymbolicStack) (size uint64, overflow bool)

type operation struct {
	execute executionFunc
	//memorySize memorySizeFunc // memorySize returns the memory size required for the operation
}

type JumpTable [256]*operation

var symbolicJumpTable JumpTable = JumpTable{
	vm.PUSH1: {
		execute: opPush1,
	},
	vm.MSTORE: {
		execute: opMstore,
	},
	vm.MSTORE8: {
		execute: opMstore8,
	},
	vm.CALLDATASIZE: {
		execute: opCallDataSize,
	},
	vm.LT: {
		execute: opLt,
	},
	vm.GT: {
		execute: opGt,
	},
	vm.SLT: {
		execute: opSlt,
	},
	vm.OR: {
		execute: opOr,
	},
	vm.XOR: {
		execute: opXor,
	},
	vm.TIMESTAMP: {
		execute: opTimestamp,
	},
	vm.BYTE: {
		execute: opByte,
	},
	vm.SHL: {
		execute: opSHL,
	},
	vm.EXTCODESIZE: {
		execute: opExtCodeSize,
	},
	vm.CALLDATACOPY: {
		execute: opCallDataCopy,
	},
	vm.CALLDATALOAD: {
		execute: opCallDataLoad,
	},
	vm.DIV: {
		execute: opDiv,
	},
	vm.AND: {
		execute: opAnd,
	},
	vm.EQ: {
		execute: opEq,
	},
	vm.JUMPDEST: {
		execute: opJumpdest,
	},
	vm.JUMP: {
		execute: opJump,
	},
	vm.STOP: {
		execute: opStop,
	},
	vm.SHR: {
		execute: opSHR,
	},
	vm.CALLVALUE: {
		execute: opCallValue,
	},
	vm.ISZERO: {
		execute: opIszero,
	},
	vm.REVERT: {
		execute: opRevert,
	},
	vm.MLOAD: {
		execute: opMload,
	},
	vm.ADD: {
		execute: opAdd,
	},
	vm.SUB: {
		execute: opSub,
	},
	vm.MUL: {
		execute: opMul,
	},
	vm.POP: {
		execute: opPop,
	},
	vm.EXP: {
		execute: opExp,
	},
	vm.NOT: {
		execute: opNot,
	},
	vm.CALLER: {
		execute: opCaller,
	},
	vm.KECCAK256: {
		execute: opKeccak256,
	},
	vm.SSTORE: {
		execute: opSstore,
	},
	vm.SLOAD: {
		execute: opSload,
	},
	vm.RETURN: {
		execute: opReturn,
	},
	vm.DUP1: {
		execute: makeDup(1),
	},
	vm.DUP2: {
		execute: makeDup(2),
	},
	vm.DUP3: {
		execute: makeDup(3),
	},
	vm.DUP4: {
		execute: makeDup(4),
	},
	vm.DUP5: {
		execute: makeDup(5),
	},
	vm.DUP6: {
		execute: makeDup(6),
	},
	vm.DUP7: {
		execute: makeDup(7),
	},
	vm.DUP8: {
		execute: makeDup(8),
	},
	vm.DUP9: {
		execute: makeDup(9),
	},
	vm.DUP10: {
		execute: makeDup(10),
	},
	vm.DUP11: {
		execute: makeDup(11),
	},
	vm.DUP12: {
		execute: makeDup(12),
	},
	vm.DUP13: {
		execute: makeDup(13),
	},
	vm.DUP14: {
		execute: makeDup(14),
	},
	vm.DUP15: {
		execute: makeDup(15),
	},
	vm.DUP16: {
		execute: makeDup(16),
	},
	vm.SWAP1: {
		execute: makeSwap(1),
	},
	vm.SWAP2: {
		execute: makeSwap(2),
	},
	vm.SWAP3: {
		execute: makeSwap(3),
	},
	vm.SWAP4: {
		execute: makeSwap(4),
	},
	vm.SWAP5: {
		execute: makeSwap(5),
	},
	vm.SWAP6: {
		execute: makeSwap(6),
	},
	vm.SWAP7: {
		execute: makeSwap(7),
	},
	vm.SWAP8: {
		execute: makeSwap(8),
	},
	vm.SWAP9: {
		execute: makeSwap(9),
	},
	vm.SWAP10: {
		execute: makeSwap(10),
	},
	vm.SWAP11: {
		execute: makeSwap(11),
	},
	vm.SWAP12: {
		execute: makeSwap(12),
	},
	vm.SWAP13: {
		execute: makeSwap(13),
	},
	vm.SWAP14: {
		execute: makeSwap(14),
	},
	vm.SWAP15: {
		execute: makeSwap(15),
	},
	vm.SWAP16: {
		execute: makeSwap(16),
	},

	vm.PUSH2: {
		execute: makePush(2, 2),
	},
	vm.PUSH3: {
		execute: makePush(3, 3),
	},
	vm.PUSH4: {
		execute: makePush(4, 4),
	},
	vm.PUSH5: {
		execute: makePush(5, 5),
	},
	vm.PUSH6: {
		execute: makePush(6, 6),
	},
	vm.PUSH7: {
		execute: makePush(7, 7),
	},
	vm.PUSH8: {
		execute: makePush(8, 8),
	},
	vm.PUSH9: {
		execute: makePush(9, 9),
	},
	vm.PUSH10: {
		execute: makePush(10, 10),
	},
	vm.PUSH11: {
		execute: makePush(11, 11),
	},
	vm.PUSH12: {
		execute: makePush(12, 12),
	},
	vm.PUSH13: {
		execute: makePush(13, 13),
	},
	vm.PUSH14: {
		execute: makePush(14, 14),
	},
	vm.PUSH15: {
		execute: makePush(15, 15),
	},
	vm.PUSH16: {
		execute: makePush(16, 16),
	},
	vm.PUSH17: {
		execute: makePush(17, 17),
	},
	vm.PUSH18: {
		execute: makePush(18, 18),
	},
	vm.PUSH19: {
		execute: makePush(19, 19),
	},
	vm.PUSH20: {
		execute: makePush(20, 20),
	},
	vm.PUSH21: {
		execute: makePush(21, 21),
	},
	vm.PUSH22: {
		execute: makePush(22, 22),
	},
	vm.PUSH23: {
		execute: makePush(23, 23),
	},
	vm.PUSH24: {
		execute: makePush(24, 24),
	},
	vm.PUSH25: {
		execute: makePush(25, 25),
	},
	vm.PUSH26: {
		execute: makePush(26, 26),
	},
	vm.PUSH27: {
		execute: makePush(27, 27),
	},
	vm.PUSH28: {
		execute: makePush(28, 28),
	},
	vm.PUSH29: {
		execute: makePush(29, 29),
	},
	vm.PUSH30: {
		execute: makePush(30, 30),
	},
	vm.PUSH31: {
		execute: makePush(31, 31),
	},
	vm.PUSH32: {
		execute: makePush(32, 32),
	},
	vm.LOG0: {
		execute: makeLog(0),
	},
	vm.LOG1: {
		execute: makeLog(1),
	},
	vm.LOG2: {
		execute: makeLog(2),
	},
	vm.LOG3: {
		execute: makeLog(3),
	},
	vm.LOG4: {
		execute: makeLog(4),
	},
	vm.GAS: {
		execute: opGas,
	},
	vm.RETURNDATASIZE: {
		execute: opReturnDataSize,
	},
	vm.RETURNDATACOPY: {
		execute: opReturnDataCopy,
	},
	vm.CODECOPY: {
		execute: opCodeCopy,
	},
	vm.ADDRESS: {
		execute: opAddress,
	},
	vm.INVALID: {
		execute: opInvalid,
	},
	vm.GASLIMIT: {
		execute: opGasLimit,
	},
	vm.GASPRICE: {
		execute: opGasPrice,
	},
	vm.NUMBER: {
		execute: opNumber,
	},
	vm.PUSH0: {
		execute: opPush0,
	},
	vm.SELFBALANCE: {
		execute: opSelfBalance,
	},
	vm.BALANCE: {
		execute: opBalance,
	},
	vm.ORIGIN: {
		execute: opOrigin,
	},
	vm.MOD: {
		execute: opMod,
	},
	vm.SGT: {
		execute: opSGt,
	},
	vm.CHAINID: {
		execute: opChainID,
	},
	vm.EXTCODEHASH: {
		execute: opExtCodeHash,
	},
	vm.SIGNEXTEND: {
		execute: opSignExtend,
	},
	vm.COINBASE: {
		execute: opCoinbase,
	},
	vm.DIFFICULTY: {
		execute: opDifficulty,
	},
	vm.BLOCKHASH: {
		execute: opBlockHash,
	},
}
