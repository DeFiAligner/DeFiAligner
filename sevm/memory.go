package sevm

import (
	"errors"
	"fmt"

	"github.com/aclements/go-z3/z3"
	"github.com/gookit/color"
)

const MEMORY_BV_SIZE = 256 * 50 // bites *  the number of slots

type SymbolicMemory struct {
	Store z3.Value
}

func NewSymbolicMemory(ctx *z3.Context) *SymbolicMemory {
	return &SymbolicMemory{
		Store: ctx.FromInt(0, ctx.BVSort(MEMORY_BV_SIZE)).(z3.BV),
	}
}

func (m *SymbolicMemory) ChangeMemoryValue(offset_uint64 uint64, size uint64, value z3.Value) {

	if size > 0 { // size must be more than 0

		memory_length := m.Store.Sort().BVSize()
		value_length := value.Sort().BVSize()
		//the slot of  start  and end points onc  memory
		start_slot := int(offset_uint64 / 32)
		end_slot := int((offset_uint64 + size - 1) / 32)

		highest_memory_value := value.Context().FromInt(0, value.Context().IntSort()).(z3.Int).ToBV(256)

		for slot_index := memory_length/256 - 1; slot_index >= 0; slot_index-- {
			current_slot_value := m.Store.(z3.BV).Extract((slot_index+1)*256-1, slot_index*256)
			//color.Error.Println(slot_index, start_slot, end_slot)

			// if there is no start_slot and end_slot
			if start_slot > slot_index || end_slot < slot_index {
				highest_memory_value = highest_memory_value.Concat(current_slot_value)
			}

			if start_slot < slot_index && end_slot > slot_index {
				have_added_length := slot_index*256 - int(offset_uint64)*8
				value_part := value.(z3.BV).Extract(value_length-have_added_length-1, value_length-have_added_length-256)
				highest_memory_value = highest_memory_value.Concat(value_part)
			}

			// if there is only end slot
			if start_slot != slot_index && end_slot == slot_index {

				current_value_length := int(offset_uint64+size)*8 - slot_index*256
				value_part := value.(z3.BV).Extract(current_value_length-1, 0)

				if current_value_length == 256 {
					highest_memory_value = highest_memory_value.Concat(value_part)
				} else {
					slot_part := current_slot_value.Extract(256-current_value_length-1, 0)
					highest_memory_value = highest_memory_value.Concat(value_part.Concat(slot_part))

				}

			}
			// if there is only start slot
			if start_slot == slot_index && end_slot != slot_index {
				current_value_length := 0
				if int(offset_uint64)*8-slot_index*256 == 0 {
					value_part := value.(z3.BV).Extract(int(size)*8-1, 256)
					highest_memory_value = highest_memory_value.Concat(value_part)
				} else {
					current_value_length = (slot_index+1)*256 - int(offset_uint64)*8
					value_part := value.(z3.BV).Extract(256-1, 256-current_value_length)
					slot_part := current_slot_value.Extract(256-1, current_value_length)
					highest_memory_value = highest_memory_value.Concat(slot_part.Concat(value_part))

				}

			}
			// if there are bot start slot and end slot
			if start_slot == slot_index && end_slot == slot_index {
				if offset_uint64%32 == 0 && (offset_uint64+size)%32 == 0 {
					value_part := value.(z3.BV).Extract(256-1, 0)
					highest_memory_value = highest_memory_value.Concat(value_part)
				}

				if offset_uint64%32 == 0 && (offset_uint64+size)%32 != 0 {
					value_part := value.(z3.BV)
					slot_part := current_slot_value.Extract(int(256-size*8-1), 0)
					highest_memory_value = highest_memory_value.Concat(value_part.Concat(slot_part))

				}
				if offset_uint64%32 != 0 && (offset_uint64+size)%32 == 0 {
					value_part := value.(z3.BV)
					slot_part := current_slot_value.Extract(256-1, int(size*8))
					highest_memory_value = highest_memory_value.Concat(slot_part.Concat(value_part))

				}

				if offset_uint64%32 != 0 && (offset_uint64+size)%32 != 0 {
					value_part := value.(z3.BV)

					first_slot_part := current_slot_value.Extract(256-1, int(offset_uint64*8-uint64(slot_index)*256-1)+int(size*8))
					second_slot_part := current_slot_value.Extract(int(offset_uint64*8-uint64(slot_index)*256-1), 0)

					// color.Error.Println(256-1, int(offset_uint64*8-uint64(slot_index)*256)+int(size*8), int(offset_uint64*8-uint64(slot_index)*256-1), 0)
					// PauseForSeconds(100)
					highest_memory_value = highest_memory_value.Concat(first_slot_part).Concat(value_part).Concat(second_slot_part)

				}

			}

		}

		highest_length := highest_memory_value.Sort().BVSize()
		final_value := highest_memory_value.Extract(highest_length-256-1, 0)
		m.Store = SimplifyZ3BV(final_value)

	}

}

func (m *SymbolicMemory) Set(offset_value z3.Value, size uint64, value z3.Value) error {

	if size > 0 {
		if value == nil {
			value = m.Store.Context().FromInt(0, m.Store.Context().IntSort()).(z3.Int).ToBV(int(size * 8))
		}

		offset_uint64, _, ok := offset_value.(z3.BV).AsUint64()
		if !ok { // can't support Z3.Value

			color.Error.Println("Memory offset is not a concrete value!")

			// color.Error.Println(offset_value)
			// PauseForSeconds(100)
			return errors.New("invalid opcode")
		}
		memory_length := m.Store.Sort().BVSize()
		// if memory is not enough, add the length
		if (offset_uint64+size)*8 > uint64(memory_length) {
			slots_num := ((offset_uint64+size)*8-uint64(memory_length))/256 + 1
			add_BV := m.Store.Context().FromInt(0, m.Store.Context().BVSort(int(slots_num*256))).(z3.BV)
			m.Store = add_BV.Concat(m.Store.(z3.BV))
		}

		if value.Sort().BVSize() < int(size)*8 {
			lacked_bv := value.Context().FromInt(0, value.Context().BVSort(int(size)*8-value.Sort().BVSize())).(z3.BV)
			value = SimplifyZ3BV(lacked_bv.Concat(value.(z3.BV)))
		}

		value = SimplifyZ3BV(value.(z3.BV).Extract(int(size*8-1), 0))
		m.ChangeMemoryValue(offset_uint64, size, value)

	}

	return nil

}

// Set sets offset + size to value
func (m *SymbolicMemory) Set32(offset_value z3.Value, value z3.Value) error {

	return m.Set(offset_value, 32, value)
}

func (m *SymbolicMemory) GetCopy(offset_value z3.Value, size uint64) z3.Value {
	offset_uint64, _, ok := offset_value.(z3.BV).AsUint64()
	if !ok { // can't support Z3.Value
		color.Error.Println("Memory offset is not a concrete value!")
		return nil
	}

	if size > 0 { // size must be more than 0

		memory_length := m.Store.Sort().BVSize()

		//the slot of  start  and end points onc  memory
		start_slot := int(offset_uint64 / 32)
		end_slot := int((offset_uint64 + size - 1) / 32)

		highest_memory_value := m.Store.Context().FromInt(0, m.Store.Context().IntSort()).(z3.Int).ToBV(256)

		for slot_index := memory_length/256 - 1; slot_index >= 0; slot_index-- {
			current_slot_value := m.Store.(z3.BV).Extract((slot_index+1)*256-1, slot_index*256)

			// // if there is no start_slot and end_slot : no action
			// if start_slot > slot_index || end_slot < slot_index {
			// 	highest_memory_value = highest_memory_value.Concat(current_slot_value)
			// }

			if start_slot < slot_index && end_slot > slot_index { //copy the value
				highest_memory_value = current_slot_value.Concat(highest_memory_value)
			}

			// if there is only end slot
			if start_slot != slot_index && end_slot == slot_index {

				current_value_length := int(offset_uint64+size)*8 - slot_index*256
				if current_value_length == 256 {
					highest_memory_value = current_slot_value.Concat(highest_memory_value)
				} else {
					slot_part := current_slot_value.Extract(256-1, 256-current_value_length)
					highest_memory_value = slot_part.Concat(highest_memory_value)
				}

			}
			// if there is only start slot
			if start_slot == slot_index && end_slot != slot_index {
				current_value_length := 0
				if int(offset_uint64)*8-slot_index*256 == 0 {
					highest_memory_value = current_slot_value.Concat(highest_memory_value)
				} else {
					current_value_length = (slot_index+1)*256 - int(offset_uint64)*8
					slot_part := current_slot_value.Extract(current_value_length-1, 0)
					highest_memory_value = slot_part.Concat(highest_memory_value)

				}

			}
			// if there are bot start slot and end slot
			if start_slot == slot_index && end_slot == slot_index {
				if offset_uint64%32 == 0 && (offset_uint64+size)%32 == 0 {
					highest_memory_value = current_slot_value.Concat(highest_memory_value)
				}

				if offset_uint64%32 == 0 && (offset_uint64+size)%32 != 0 {
					slot_part := current_slot_value.Extract(256-1, 256-int(size)*8)
					highest_memory_value = slot_part.Concat(highest_memory_value)

				}
				if offset_uint64%32 != 0 && (offset_uint64+size)%32 == 0 {
					slot_part := current_slot_value.Extract(int(size*8-1), 0)
					highest_memory_value = slot_part.Concat(highest_memory_value)

				}

				if offset_uint64%32 != 0 && (offset_uint64+size)%32 != 0 {
					his_pos := 256 - (offset_uint64*8 - uint64(slot_index)*256)
					slot_part := current_slot_value.Extract(int(his_pos-1), int(his_pos-size*8))
					highest_memory_value = slot_part.Concat(highest_memory_value)

				}

			}

		}

		highest_length := highest_memory_value.Sort().BVSize()

		final_value := highest_memory_value.Extract(highest_length-1, 256)
		return SimplifyZ3BV(final_value)

	}
	return nil

}

// // GetCopy returns offset + size as a new slice
// func (m *SymbolicMemory) GetCopy66(offset z3.Value, size uint64) z3.Value {

// 	offset_uint64, _, ok := offset.(z3.BV).AsUint64()
// 	if !ok { // can't support Z3.Value
// 		color.Error.Println("Memory offset is not a concrete value!")
// 		os.Exit(1)
// 	}

// 	memory_length := m.Store.Sort().BVSize()
// 	if size == 0 {
// 		return nil
// 	}
// 	// if memory is not enough, add the length
// 	if (offset_uint64+size)*8 > uint64(memory_length) {
// 		slots_num := ((offset_uint64+size)*8-uint64(memory_length))/256 + 1
// 		add_BV := m.Store.Context().FromInt(0, m.Store.Context().BVSort(int(slots_num*256))).(z3.BV)
// 		m.Store = add_BV.Concat(m.Store.(z3.BV))
// 	}

// 	// if  the copy is in the same slot
// 	var value z3.BV
// 	if offset_uint64%32 == 0 && size == 32 {
// 		value = m.Store.(z3.BV).Extract(int(offset_uint64*8+size*8-1), int(offset_uint64*8))
// 	}

// 	// if  the copy is in two slots and the leght is less than 64
// 	if (size > 32 && size <= 64) || (offset_uint64%32 != 0 && size <= 64) {
// 		first_length := (offset_uint64/32+1)*256 - offset_uint64*8
// 		second_length := size*8 - first_length
// 		first_part := m.Store.(z3.BV).Extract(int((offset_uint64/32)*256+first_length-1), int((offset_uint64/32)*256))
// 		second_part := m.Store.(z3.BV).Extract(int((offset_uint64/32+2)*256-1), int((offset_uint64/32+2)*256-second_length))
// 		value = first_part.Concat(second_part)

// 	}
// 	// 	if  the copy is in  more slots and the leght is more than 64

// 	if size > 64 {
// 		if offset_uint64%32 == 0 && (offset_uint64+size)%32 == 0 {
// 			value = m.Store.(z3.BV).Extract(int(offset_uint64*8+size*8-1), int(offset_uint64*8))
// 		}
// 		if offset_uint64%32 != 0 && (offset_uint64+size)%32 == 0 {
// 			first_length := (offset_uint64/32+1)*256 - offset_uint64*8
// 			//second_length := size*8 - first_length
// 			first_part := m.Store.(z3.BV).Extract(int((offset_uint64/32)*256+first_length-1), int((offset_uint64/32)*256))
// 			second_part := m.Store.(z3.BV).Extract(int(offset_uint64*8+size*8-1), int((offset_uint64/32)*256+first_length))
// 			value = first_part.Concat(second_part)

// 		}
// 		if offset_uint64%32 == 0 && (offset_uint64+size)%32 != 0 {
// 			first_length := size / 32 * 256
// 			second_length := size*8 - first_length
// 			first_part := m.Store.(z3.BV).Extract(int(offset_uint64*8+first_length-1), int(offset_uint64*8))
// 			second_part := m.Store.(z3.BV).Extract(int(((offset_uint64+size)/32+1)*256-1), int(((offset_uint64+size)/32+1)*256-second_length))
// 			value = first_part.Concat(second_part)
// 		}
// 		if offset_uint64%32 != 0 && (offset_uint64+size)%32 != 0 {
// 			first_length := (offset_uint64/32+1)*256 - offset_uint64*8
// 			third_length := offset_uint64*8 + size*8 - (offset_uint64+size)/32*256
// 			second_length := size*8 - first_length - third_length

// 			first_part := m.Store.(z3.BV).Extract(int((offset_uint64/32)*256+first_length-1), int((offset_uint64/32)*256))

// 			second_part := m.Store.(z3.BV).Extract(int((offset_uint64/32+1)*256+second_length-1), int((offset_uint64/32+1)*256))

// 			third_part := m.Store.(z3.BV).Extract(int(((offset_uint64+size)/32+1)*256-1), int(((offset_uint64+size)/32)*256+256-third_length))

// 			value = first_part.Concat(second_part).Concat(third_part)

// 		}

// 	}

// 	return SimplifyZ3BV(value)

// }

func (m *SymbolicMemory) PrintSymbolicMemory() {
	color.Warn.Println("**************** Memory Start *****************")
	memory_length := m.Store.Sort().BVSize()

	for i := 0; i < memory_length/256; i += 1 {
		offset := i * 32
		size := 32
		value := m.Store.Context().Simplify(m.Store.(z3.BV).Extract(i*256+256-1, i*256), m.Store.Context().Config())
		offset_BV := m.Store.Context().FromInt(int64(offset), m.Store.Context().BVSort(32)).(z3.BV)
		formattedStr := fmt.Sprintf("Slot: %-4d Offset: %s Size: %-4d Value: %s", i, offset_BV, size, value)
		color.Info.Println(formattedStr)
	}
	color.Warn.Println("**************** Memory End *****************")
}

func (m *SymbolicMemory) Len() int {
	return m.Store.Sort().BVSize()
}

func (m *SymbolicMemory) Copy() SymbolicMemory {
	copid_momory := NewSymbolicMemory(m.Store.Context())
	copid_momory.Store = m.Store
	return *copid_momory
}
