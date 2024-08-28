package sevm

import (
	"fmt"
	"sync"

	"reflect"

	"github.com/aclements/go-z3/z3"
	"github.com/gookit/color"
	// "github.com/gookit/color"
)

// The pool is used to cache and reuse SymbolicStack objects in order to reduce the overhead of object allocation and deallocation, thereby enhancing program performance.
var stackPool = sync.Pool{
	New: func() interface{} {
		return &SymbolicStack{data: make([]z3.Value, 0, 16)}
	},
}

type SymbolicStack struct {
	data []z3.Value
}

func newstack() *SymbolicStack {
	return stackPool.Get().(*SymbolicStack)
}

func returnStack(s *SymbolicStack) {
	s.data = s.data[:0]
	stackPool.Put(s)
}

func (st *SymbolicStack) Data() []z3.Value {
	return st.data
}

func (st *SymbolicStack) push(d z3.Value) {
	// Push the element to the stack
	st.data = append(st.data, d)
}

func (st *SymbolicStack) pop() (ret z3.Value) {
	// Pop the element from the stack
	if len(st.data) > 0 {
		ret = st.data[len(st.data)-1]
		st.data = st.data[:len(st.data)-1]
	}
	return
}

func (st *SymbolicStack) len() int {
	return len(st.data)
}

func (st *SymbolicStack) swap(n int) {
	st.data[st.len()-n], st.data[st.len()-1] = st.data[st.len()-1], st.data[st.len()-n]
}

func (st *SymbolicStack) dup(n int) {
	st.push(st.data[st.len()-n])
}

// func (st *SymbolicStack) peek() z3.Value {
// 	if len(st.data) > 0 {
// 		return st.data[len(st.data)-1]
// 	}
// 	return nil
// }

// Back returns the n'th item in stack
func (st *SymbolicStack) Back(n int) z3.Value {
	return st.data[st.len()-n-1]
}

func isSimilarWithinThreshold(str1, str2 string) bool {
	diffCount := 0
	length := len(str1)

	for i := 0; i < length; i++ {
		if str1[i] != str2[i] {
			diffCount++
		}
	}
	diffPercentage := float64(diffCount) / float64(length)

	return diffPercentage <= 0.15
}

func AreSymbolicEqual(value1 z3.Value, value2 z3.Value) bool {
	//Note: value 1 and value2 must be  Int or BV

	if value1.String() == value2.String() {
		return true
	}
	// value1 = value1.(z3.BV).SToInt()
	// value2 = value2.(z3.BV).SToInt()

	value1 = SimplifyZ3BV(value1.(z3.BV))
	value2 = SimplifyZ3BV(value2.(z3.BV))

	if len(value1.String()) != len(value2.String()) {
		return false
	}

	return isSimilarWithinThreshold(value1.String(), value2.String())

	//return value1.String() == value2.String()

	// ctx := value1.Context()
	// solver := z3.NewSolver(ctx)

	// notEqual := value1.(z3.BV).NE(value2.(z3.BV))

	// solver.Assert(notEqual)
	// result, _ := solver.Check()

	// return result

	//return false

	// solver1 := z3.NewSolver(value1.Context())
	// solver2 := z3.NewSolver(value1.Context())
	// solver3 := z3.NewSolver(value1.Context())
	// int_value := value1.Context().IntSort()

	// eq_zero := value1.(z3.Int).Sub(value2.(z3.Int)).Eq(value1.Context().FromInt(0, int_value).(z3.Int)) // value1-value2==0 have a solution?
	// gt_zero := value1.(z3.Int).Sub(value2.(z3.Int)).GT(value1.Context().FromInt(0, int_value).(z3.Int)) // value1-value2>0 have a solution?
	// lt_zero := value1.(z3.Int).Sub(value2.(z3.Int)).LT(value1.Context().FromInt(0, int_value).(z3.Int)) // value1-value2<0 have a solution?

	// solver1.Assert(eq_zero)
	// if_solved1, _ := solver1.Check()
	// solver2.Assert(gt_zero)
	// if_solved2, _ := solver2.Check()
	// solver3.Assert(lt_zero)
	// if_solved3, _ := solver3.Check()

	// if if_solved1 && !if_solved2 && !if_solved3 { // only have a solution for EQ
	// 	return true
	// } else {
	// 	return false
	// }

}

func AreSymbolicStacksEqual(stack1 *SymbolicStack, stack2 *SymbolicStack) bool {
	//Note : stack1, stack2  must be under the same z3 config and contex
	if len(stack1.data) != len(stack2.data) {
		return false
	}
	//
	//tolerances := 0
	for i := 0; i < len(stack1.data); i++ {
		value1 := stack1.data[i]
		value2 := stack2.data[i]
		if IsConcreteValue(value1.(z3.BV)) != IsConcreteValue(value2.(z3.BV)) { // if they are not the same type
			return false
		}
		if !AreSymbolicEqual(value1, value2) {
			return false
		}
		// if tolerances >= 1 { //
		// 	return false
		// }

	}
	return true
}

func (stack *SymbolicStack) PrintCurrentStack() {
	color.Warn.Println("---------------- Stack Content ----------------")
	for i := len(stack.data) - 1; i >= 0; i-- {
		val := stack.data[i]
		valType := reflect.TypeOf(val)
		//fmt.Printf("[%d]: %s     %+v\n", len(stack.data)-i, valType.Name(), val)
		formattedStr := fmt.Sprintf("Number: [%-2d] Type: %s Value: %s", len(stack.data)-i, valType.Name(), val)
		color.Info.Println(formattedStr)
	}
	color.Warn.Println("---------------- End of Stack ----------------")
}

func (st *SymbolicStack) Len() int {
	return len(st.data)
}

func (s *SymbolicStack) Copy() *SymbolicStack {
	copyStack := &SymbolicStack{
		data: make([]z3.Value, len(s.data)),
	}
	copy(copyStack.data, s.data)
	return copyStack
}
