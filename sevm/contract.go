package sevm

import (
	"github.com/aclements/go-z3/z3"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	// "github.com/holiman/uint256"
)

// ContractRef is a reference to the contract's backing object
type ContractRef interface {
	Address() common.Address
}

// AccountRef implements ContractRef.
//
// Account references are used during EVM initialisation and
// its primary use is to fetch addresses. Removing this object
// proves difficult because of the cached jump destinations which
// are fetched from the parent contract (i.e. the caller), which
// is a ContractRef.

// Contract represents an ethereum contract in the state database. It contains
// the contract code, calling arguments. Contract implements ContractRef
type Contract struct {
	// CallerAddress is the result of the caller which initialised this
	// contract. However when the "call method" is delegated this value
	// needs to be initialised to that of the caller's caller.
	// CallerAddress z3.BV //
	caller z3.BV // who calls the contract
	self   z3.BV // the address of the contract

	// jumpdests map[common.Hash]bitvec // Aggregated result of JUMPDEST analysis.
	// analysis  bitvec                 // Locally cached result of JUMPDEST analysis
	Code     []byte
	CodeHash z3.BV
	CodeAddr z3.BV
	Input    z3.BV

	Gas   uint64
	value z3.BV // ETH Value
}

// NewContract returns a new contract environment for the execution of EVM.
func NewContract(caller z3.BV, self_address z3.BV, value z3.BV, gas uint64) *Contract {
	c := &Contract{caller: caller, self: self_address}

	// if parent, ok := caller.(*Contract); ok {
	// 	// Reuse JUMPDEST analysis from parent context if available.
	// 	c.jumpdests = parent.jumpdests
	// } else {
	// 	c.jumpdests = make(map[common.Hash]bitvec)
	// }

	// Gas should be a pointer so it can safely be reduced through the run
	// This pointer will be off the state transition
	c.Gas = gas
	// ensures a value is set
	c.value = value

	return c
}

// GetOp returns the n'th element in the contract's byte array
func (c *Contract) GetOp(n uint64) vm.OpCode {
	if n < uint64(len(c.Code)) {
		return vm.OpCode(c.Code[n])
	}

	return vm.STOP
}

// Caller returns the caller of the contract.
//
// Caller will recursively call caller when the contract is a delegate
// call, including that of caller's caller.
func (c *Contract) Caller() z3.Value {
	return c.caller
}

func (c *Contract) validJumpdest(dest uint64) bool {

	// PC cannot go beyond len(code) and certainly can't be bigger than 63bits.
	// Don't bother checking for JUMPDEST in that case.
	if dest >= uint64(len(c.Code)) {
		return false
	}
	// Only JUMPDESTs allowed for destinations
	if vm.OpCode(c.Code[dest]) != vm.JUMPDEST {
		return false
	}

	return true
	// return c.isCode(udest)
}

func (c *Contract) SetCallCode(addr z3.BV, code []byte) {
	c.Code = code
	// c.CodeHash = hash
	c.CodeAddr = addr
}

func (c *Contract) Copy() *Contract {
	copyContract := *c

	if c.Code != nil {
		copyContract.Code = make([]byte, len(c.Code))
		copy(copyContract.Code, c.Code)
	}
	return &copyContract
}
