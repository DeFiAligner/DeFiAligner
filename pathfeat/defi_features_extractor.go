package pathfeat

//Extract key DeFi information from the path.

import (
	"github.com/DeFiAligner/DeFiAligner/sevm"
	"github.com/aclements/go-z3/z3"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/sergi/go-diff/diffmatchpatch"
	"regexp"
	"strings"
)

// func (dc DeFiCondictionList) DeepCopy() DeFiCondictionList {
// 	copyCond := make(DeFiCondictionList, len(dc))
// 	copy(copyCond, dc)
// 	return copyCond
// }

func IsElementInCondictList(element z3.Bool, condiction_list []z3.Bool) bool {
	for _, item := range condiction_list {
		if element.String() == item.String() {
			return true
		}
	}
	return false
}

type EthTransferInfo struct {
	Sender z3.BV
	Payee  z3.BV
	Value  z3.BV
}

// Balance Change Info
type BalanceChangeInfo struct {
	SstorageChangeLocation z3.BV
	SstorageChangeValue    z3.BV
}

// Map Address to BalanceChange  List
type AddressToBalanceChanges map[string][]BalanceChangeInfo
type AddressToCondictions map[string][]z3.BV
type TokenBalanceSymbol map[string]z3.BV

// DeFiFeature
type DeFiFeature struct {
	ERCBalanceChangeMap AddressToBalanceChanges
	CondictionMap       AddressToCondictions
	EthTransfer         []EthTransferInfo
}

func isHex(s string) bool {
	re := regexp.MustCompile(`^(0x|0X)?[0-9a-fA-F]+$`)
	return re.MatchString(s)
}

// compare balance string
func IsSameBlanceStruct(sstore_string, balance_string string) bool {

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(sstore_string, balance_string, false)
	sstore_diff_string := ""
	delete_count := 0
	insert_count := 0
	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			{
				insert_count += 1
				if insert_count > 1 {
					return false
				}

			}
		case diffmatchpatch.DiffDelete:
			sstore_diff_string = diff.Text
			delete_count += 1
			if delete_count > 1 {
				return false
			}
		case diffmatchpatch.DiffEqual:
			{
				// pass
			}
		}
	}

	if delete_count == 1 && insert_count == 1 && !isHex(strings.Replace(sstore_diff_string, "#x", "", -1)) { // if there is only one difference and its not a HEX.
		return true
	}

	return false

}

func GetDeFiSymbolicFeaturesFromPath(path sevm.ExecutionPath, called_contract z3.BV) DeFiFeature {

	balance_change_map := AddressToBalanceChanges{}
	balance_symbol_map := TokenBalanceSymbol{}
	var eth_transfers []EthTransferInfo
	zero_value := sevm.SimplifyZ3BV(called_contract.Context().FromInt(0, called_contract.Context().IntSort()).(z3.Int).ToBV(256))
	one_value := sevm.SimplifyZ3BV(called_contract.Context().FromInt(1, called_contract.Context().IntSort()).(z3.Int).ToBV(256))
	address_to_conditions_map := AddressToCondictions{} // dynamic condition
	current_called_contract := called_contract

	for index, executionState := range path {

		// get BalanceSymbol
		current_called_address_string := "0x" + current_called_contract.String()[26:66]
		_, exists := balance_symbol_map[current_called_address_string]
		if !exists {
			balance_symbol_list := GetBalanceSymbol(current_called_address_string, called_contract.Context())
			if len(balance_symbol_list) == 1 { //  By default, there is only one balance expression for the balance
				balance_symbol_map[current_called_address_string] = balance_symbol_list[0]
			}
		}

		// update JUMPI condictions
		if executionState.Current_opcode == vm.JUMPI {
			current_pc := executionState.Current_pc
			next_loaction := path[index+1].Current_pc
			var jumpi_condition z3.Bool
			condiction := sevm.SimplifyZ3BV(path[index-1].Current_stack.Back(1).(z3.BV))

			if next_loaction == current_pc+1 {
				jumpi_condition = condiction.Eq(zero_value)
			} else {
				jumpi_condition = condiction.Eq(one_value)
			}

			jumpi_condition = jumpi_condition.Context().Simplify(jumpi_condition, jumpi_condition.Context().Config()).(z3.Bool)
			_, is_concrete_condiction := jumpi_condition.AsBool()

			if !is_concrete_condiction {
				address_to_conditions_map[current_called_address_string] = append(address_to_conditions_map[current_called_address_string], z3.BV(jumpi_condition))
			}

		}

		// update the called contract: if a new contract is called, update the called contract by checking  the stack
		if executionState.Current_opcode == vm.CALL || executionState.Current_opcode == vm.DELEGATECALL || executionState.Current_opcode == vm.STATICCALL { //staticcall will not change sstorage
			current_called_contract = path[index-1].Current_stack.Back(1).(z3.BV)

			// collect ETH transfer
			if executionState.Current_opcode == vm.CALL && path[index-1].Current_stack.Back(2).String() != "#x0000000000000000000000000000000000000000000000000000000000000000" {
				eth_transfer := EthTransferInfo{}
				eth_transfer.Sender = path[index-1].Current_called_contract
				eth_transfer.Payee = path[index-1].Current_stack.Back(1).(z3.BV)
				eth_transfer.Value = path[index-1].Current_stack.Back(2).(z3.BV)
				eth_transfers = append(eth_transfers, eth_transfer)
			}

		}
		//  If there is a RETURN or STOP, update the called contract
		if executionState.Current_opcode == vm.RETURN || executionState.Current_opcode == vm.STOP {
			if index+1 < len(path) {
				if path[index].Current_evm_depth == path[index+1].Current_evm_depth { // if it is still in the same contract with  different functtions
					current_called_contract = path[index].Current_called_contract
				} else {
					current_called_contract = path[index+1].Current_called_contract
				}

			}
		}

		// when the balance changes , record the loc, value, and condictions of users (We do not consider the application's address, because an application may have multiple addresses.)
		if executionState.Current_opcode == vm.SSTORE {
			loc := path[index-1].Current_stack.Back(0).(z3.BV)
			if strings.Contains(loc.String(), "SHA3") {
				sstore_location_string := strings.Replace("|SLOAD"+current_called_contract.String()+"=>"+loc.String(), "|", "", -1)
				token_balance_string := strings.Replace(balance_symbol_map[current_called_address_string].String(), "|", "", -1)
				if IsSameBlanceStruct(sstore_location_string, token_balance_string) {

					balance_change_info := BalanceChangeInfo{
						SstorageChangeLocation: loc,
						SstorageChangeValue:    path[index-1].Current_stack.Back(1).(z3.BV),
					}

					balance_change_map[current_called_address_string] = append(balance_change_map[current_called_address_string], balance_change_info)

				}
			}

		}
	}

	return DeFiFeature{ERCBalanceChangeMap: balance_change_map, EthTransfer: eth_transfers, CondictionMap: address_to_conditions_map}

}
