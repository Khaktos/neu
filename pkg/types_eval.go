package neu

import "fmt"

type BaseValueType int

const (
	Nul_type BaseValueType = iota
	Num_type
	Str_type
	Bool_type
	Char_type
	List_type
)

var NoVar = Variable{}

type ValueType struct {
	BaseType BaseValueType
	SubType  []ValueType
}

func (v ValueType) String() string {
	switch v.BaseType {
	case Nul_type:
		return "NIL_T"
	case Num_type:
		return "NUM_T"
	case Str_type:
		return "STR_T"
	case Bool_type:
		return "BOOL_T"
	case Char_type:
		return "CHAR_T"
	case List_type:
		return "LIST_T[" + v.SubType[0].String() + "]"
	default:
		return fmt.Sprintf("%d", v.BaseType)
	}
}

// Base types
var BaseNum = ValueType{Num_type, nil}
var BaseStr = ValueType{Str_type, nil}
var BaseBool = ValueType{Bool_type, nil}
var BaseChar = ValueType{Char_type, nil}

var token_to_type = map[TokenType]ValueType{
	Type_num:  BaseNum,
	Type_str:  BaseStr,
	Type_bool: BaseBool,
	Type_char: BaseChar,
}

type Variable struct {
	Id      string
	Valtype ValueType
}

func isSameType(a, b ValueType) bool {
	if a.BaseType != b.BaseType {
		return false
	}
	if len(a.SubType) != len(b.SubType) {
		return false
	}
	for idx, sub_a := range a.SubType {
		if !isSameType(sub_a, b.SubType[idx]) {
			return false
		}
	}
	return true
}

type Val interface {
	Get() any
}
type Comparable interface {
	Val
	Equal(Comparable) Logical
}

type Indexable interface {
	Index(IntVal) (Val, error)
}

type Addable interface {
	Comparable
	Add(Addable) (Addable, error)
}
type Logical interface {
	Comparable
	Not()
	And(Logical)
	Or(Logical)
}
type Ordered interface {
	Comparable
	Greater(Ordered) (Logical, error)
}
type Numeric interface {
	Addable
	Ordered
	Invert() error
	Subtract(Numeric) (Numeric, error)
	Multiply(Numeric) (Numeric, error)
	Divide(Numeric) (Numeric, error)
	IntDiv(Numeric) (Numeric, error)
	Modulo(Numeric) (Numeric, error)
}

type IntVal struct {
	Value int
}
type FloatVal struct {
	Value float64
}
type StrVal struct {
	Value string
}
type CharVal struct {
	Value rune
}
type BoolVal struct {
	Value bool
}
type NullVal struct {
}

type ListVal struct {
	Stored []Val
}

func (n *NullVal) Get() any {
	return nil
}
func (i *IntVal) Get() any {
	return i.Value
}
func (f *FloatVal) Get() any {
	return f.Value
}
func (s *StrVal) Get() any {
	return s.Value
}
func (c *CharVal) Get() any {
	return c.Value
}
func (b *BoolVal) Get() any {
	return b.Value
}
func (l *ListVal) Get() any {
	return l.Stored
}

func (i *BoolVal) And(other Logical) {
	i.Value = i.Value && other.Get().(bool)
}
func (i *BoolVal) Or(other Logical) {
	i.Value = i.Value || other.Get().(bool)
}
func (i *BoolVal) Not() {
	i.Value = !i.Value
}

func (i *BoolVal) Equal(other Comparable) Logical {
	inner, ok := other.(*BoolVal)
	i.Value = ok && i.Value == inner.Value
	return i
}
func (i *StrVal) Equal(other Comparable) Logical {
	inner, ok := other.(*StrVal)
	return &BoolVal{ok && i.Value == inner.Value}
}
func (i *IntVal) Equal(other Comparable) Logical {
	inner_i, isint := other.(*IntVal)
	_, isfloat := other.(*FloatVal)
	return &BoolVal{isint && !isfloat && i.Value == inner_i.Value}
}
func (i *FloatVal) Equal(other Comparable) Logical {
	_, isint := other.(*IntVal)
	inner_f, isfloat := other.(*FloatVal)
	return &BoolVal{!isint && isfloat && i.Value == inner_f.Value}
}
func (i *CharVal) Equal(other Comparable) Logical {
	inner, ok := other.(*CharVal)
	return &BoolVal{ok && i.Value == inner.Value}
}

func (i *StrVal) Add(other Addable) (Addable, error) {
	inner, ok := other.(*StrVal)
	if !ok {
		return nil, fmt.Errorf("Not a string")
	}
	i.Value = fmt.Sprint(i.Value, inner.Value)
	return i, nil
}
func (i *IntVal) Add(other Addable) (Addable, error) {
	inner_i, isint := other.(*IntVal)
	inner_f, isfloat := other.(*FloatVal)
	if !isint && !isfloat {
		return nil, fmt.Errorf("Not a number")
	}
	if isint {
		i.Value = i.Value + inner_i.Value
		return i, nil
	}
	return &FloatVal{float64(i.Value) + inner_f.Value}, nil
}
func (i *FloatVal) Add(other Addable) (Addable, error) {
	inner_i, isint := other.(*IntVal)
	inner_f, isfloat := other.(*FloatVal)
	if !isint && !isfloat {
		return nil, fmt.Errorf("Not a number")
	}
	if isfloat {
		i.Value = i.Value + inner_f.Value
		return i, nil
	}
	return &FloatVal{i.Value + float64(inner_i.Value)}, nil
}

func (i *IntVal) Greater(other Ordered) (Logical, error) {
	inner_i, isint := other.(*IntVal)
	inner_f, isfloat := other.(*FloatVal)
	if !isint && !isfloat {
		return nil, fmt.Errorf("Not a number")
	}
	if isint {
		return &BoolVal{i.Value > inner_i.Value}, nil
	}
	return &BoolVal{float64(i.Value) > inner_f.Value}, nil
}
func (i *FloatVal) Greater(other Ordered) (Logical, error) {
	inner_i, isint := other.(*IntVal)
	inner_f, isfloat := other.(*FloatVal)
	if !isint && !isfloat {
		return nil, fmt.Errorf("Not a number")
	}
	if isfloat {
		return &BoolVal{i.Value > inner_f.Value}, nil
	}
	return &BoolVal{i.Value > float64(inner_i.Value)}, nil
}

func (i *IntVal) Invert() error {
	i.Value = -i.Value
	return nil
}
func (i *IntVal) Subtract(other Numeric) (Numeric, error) {
	inner_i, isint := other.(*IntVal)
	inner_f, isfloat := other.(*FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isint {
		i.Value = i.Value - inner_i.Value
		return i, nil
	}
	return &FloatVal{float64(i.Value) - inner_f.Value}, nil
}
func (i *IntVal) Multiply(other Numeric) (Numeric, error) {
	panic("TODO")
}
func (i *IntVal) Divide(other Numeric) (Numeric, error) {
	panic("TODO")
}
func (i *IntVal) IntDiv(other Numeric) (Numeric, error) {
	panic("TODO")
}
func (i *IntVal) Modulo(other Numeric) (Numeric, error) {
	panic("TODO")
}
func (i *FloatVal) Invert() error {
	i.Value = -i.Value
	return nil
}
func (i *FloatVal) Subtract(other Numeric) (Numeric, error) {
	inner_i, isint := other.(*IntVal)
	inner_f, isfloat := other.(*FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isfloat {
		i.Value = i.Value - inner_f.Value
		return i, nil
	}
	return &FloatVal{i.Value - float64(inner_i.Value)}, nil
}
func (i *FloatVal) Multiply(other Numeric) (Numeric, error) {
	panic("TODO")
}
func (i *FloatVal) Divide(other Numeric) (Numeric, error) {
	panic("TODO")
}
func (i *FloatVal) IntDiv(other Numeric) (Numeric, error) {
	panic("TODO")
}
func (i *FloatVal) Modulo(other Numeric) (Numeric, error) {
	panic("TODO")
}

func binCompat[T Val](a, b Val, oper string) (T, T, error) {
	ia, aok := a.(T)
	ib, bok := b.(T)
	if !aok || !bok {
		return ia, ib, fmt.Errorf("%s binary operation not defined between %T and %T", oper, a, b)
	}
	return ia, ib, nil
}
func unCompat[T any](a any, oper string) (T, error) {
	ia, aok := a.(T)
	if !aok {
		return ia, fmt.Errorf("%s unary operation not defined on %T", oper, a)
	}
	return ia, nil
}

func OpEqual(a, b Val) (Logical, error) {
	ia, ib, err := binCompat[Comparable](a, b, "Equal")
	if err != nil {
		return nil, err
	}
	return (ia).Equal(ib), nil
}
func OpNotEqual(a, b Val) (Logical, error) {
	ia, ib, err := binCompat[Comparable](a, b, "NotEqual")
	if err != nil {
		return nil, err
	}
	temp := (ia).Equal(ib)
	temp.Not()
	return temp, nil
}

func OpNot(a Val) (Logical, error) {
	ia, err := unCompat[Logical](a, "Not")
	(ia).Not()
	return ia, err
}
func OpAnd(a, b Val) (Logical, error) {
	ia, ib, err := binCompat[Logical](a, b, "And")
	if err != nil {
		return nil, err
	}
	(ia).And(ib)
	return ia, err
}
func OpOr(a, b Val) (Logical, error) {
	ia, ib, err := binCompat[Logical](a, b, "Or")
	if err != nil {
		return nil, err
	}
	(ia).Or(ib)
	return ia, err
}
func OpAdd(a, b Val) (Addable, error) {
	ia, ib, err := binCompat[Addable](a, b, "Addition")
	if err != nil {
		return nil, err
	}
	return (ia).Add(ib)
}
func OpGreater(a, b Val) (Logical, error) {
	ia, ib, err := binCompat[Ordered](a, b, "Greater")
	if err != nil {
		return nil, err
	}
	return (ia).Greater(ib)
}
func OpLess(a, b Val) (Logical, error) {
	ia, ib, err := binCompat[Ordered](a, b, "Less")
	if err != nil {
		return nil, err
	}
	step1, _ := (ia).Greater(ib)
	step2 := (ia).Equal(ib)
	step1.Not()
	step2.Not()
	step1.And(step2)
	return step1, nil
}
func OpGrEqual(a, b Val) (Logical, error) {
	ia, ib, err := binCompat[Ordered](a, b, "GrEqual")
	if err != nil {
		return nil, err
	}
	step1, _ := (ia).Greater(ib)
	step2 := (ia).Equal(ib)
	step1.And(step2)
	return step1, nil

}
func OpLeEqual(a, b Val) (Logical, error) {
	ia, ib, err := binCompat[Ordered](a, b, "LeEqual")
	if err != nil {
		return nil, err
	}
	step1, _ := (ia).Greater(ib)
	step1.Not()
	return step1, nil
}

func OpInvert(a any) (Numeric, error) {
	ia, err := unCompat[Numeric](a, "Invert")
	if err != nil {
		return nil, err
	}
	(ia).Invert()
	return ia, nil
}
func OpSubtract(a, b Val) (Numeric, error) {
	ia, ib, err := binCompat[Numeric](a, b, "Subtract")
	if err != nil {
		return nil, err
	}
	return (ia).Subtract(ib)
}
func OpMultiply(a, b Val) (Numeric, error) {
	ia, ib, err := binCompat[Numeric](a, b, "Multiply")
	if err != nil {
		return nil, err
	}
	return (ia).Multiply(ib)
}
func OpDivide(a, b Val) (Numeric, error) {
	ia, ib, err := binCompat[Numeric](a, b, "Divide")
	if err != nil {
		return nil, err
	}
	return (ia).Divide(ib)
}
func OpIntDiv(a, b Val) (Numeric, error) {
	ia, ib, err := binCompat[Numeric](a, b, "IntDiv")
	if err != nil {
		return nil, err
	}
	return (ia).IntDiv(ib)
}
func OpModulo(a, b Val) (Numeric, error) {
	ia, ib, err := binCompat[Numeric](a, b, "Modulo")
	if err != nil {
		return nil, err
	}
	return (ia).Modulo(ib)
}
