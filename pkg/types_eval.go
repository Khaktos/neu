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
	Name() string
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
	Not() Logical
	And(Logical) Logical
	Or(Logical) Logical
}
type Ordered interface {
	Comparable
	Greater(Ordered) (Logical, error)
}
type Numeric interface {
	Addable
	Ordered
	Invert() (Numeric, error)
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

func (n NullVal) Get() any {
	return nil
}
func (i IntVal) Get() any {
	return i.Value
}
func (f FloatVal) Get() any {
	return f.Value
}
func (s StrVal) Get() any {
	return s.Value
}
func (c CharVal) Get() any {
	return c.Value
}
func (b BoolVal) Get() any {
	return b.Value
}
func (l ListVal) Get() any {
	return l.Stored
}
func (n NullVal) Name() string {
	return "SEMMI"
}
func (i IntVal) Name() string {
	return "NUM"
}
func (f FloatVal) Name() string {
	return "NUM"
}
func (s StrVal) Name() string {
	return "TXT"
}
func (c CharVal) Name() string {
	return "KAR"
}
func (b BoolVal) Name() string {
	return "LOG"
}
func (l ListVal) Name() string {
	return "LIST"
}

func (i BoolVal) And(other Logical) Logical {
	return BoolVal{i.Value && other.Get().(bool)}
}
func (i BoolVal) Or(other Logical) Logical {
	return BoolVal{i.Value || other.Get().(bool)}
}
func (i BoolVal) Not() Logical {
	return BoolVal{!i.Value}
}

func (i BoolVal) Equal(other Comparable) Logical {
	inner, ok := other.(BoolVal)
	return BoolVal{ok && i.Value == inner.Value}
}
func (i StrVal) Equal(other Comparable) Logical {
	inner, ok := other.(StrVal)
	return BoolVal{ok && i.Value == inner.Value}
}
func (i IntVal) Equal(other Comparable) Logical {
	inner_i, isint := other.(IntVal)
	_, isfloat := other.(FloatVal)
	return BoolVal{isint && !isfloat && i.Value == inner_i.Value}
}
func (i FloatVal) Equal(other Comparable) Logical {
	_, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	return BoolVal{!isint && isfloat && i.Value == inner_f.Value}
}
func (i CharVal) Equal(other Comparable) Logical {
	inner, ok := other.(CharVal)
	return BoolVal{ok && i.Value == inner.Value}
}

func (i StrVal) Add(other Addable) (Addable, error) {
	inner, ok := other.(StrVal)
	if !ok {
		return nil, fmt.Errorf("Not a string")
	}
	return StrVal{fmt.Sprint(i.Value, inner.Value)}, nil
}
func (i IntVal) Add(other Addable) (Addable, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		return nil, fmt.Errorf("Not a number")
	}
	if isint {
		return IntVal{i.Value + inner_i.Value}, nil
	}
	return FloatVal{float64(i.Value) + inner_f.Value}, nil
}
func (i FloatVal) Add(other Addable) (Addable, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		return nil, fmt.Errorf("Not a number")
	}
	if isfloat {
		return FloatVal{i.Value + inner_f.Value}, nil
	}
	return FloatVal{i.Value + float64(inner_i.Value)}, nil
}

func (i IntVal) Greater(other Ordered) (Logical, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		return nil, fmt.Errorf("Not a number")
	}
	if isint {
		return BoolVal{i.Value > inner_i.Value}, nil
	}
	return BoolVal{float64(i.Value) > inner_f.Value}, nil
}
func (i FloatVal) Greater(other Ordered) (Logical, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		return nil, fmt.Errorf("Not a number")
	}
	if isfloat {
		return BoolVal{i.Value > inner_f.Value}, nil
	}
	return BoolVal{i.Value > float64(inner_i.Value)}, nil
}

func (i IntVal) Invert() (Numeric, error) {
	return IntVal{-i.Value}, nil
}
func (i IntVal) Subtract(other Numeric) (Numeric, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isint {
		return IntVal{i.Value - inner_i.Value}, nil
	}
	return FloatVal{float64(i.Value) - inner_f.Value}, nil
}
func (i IntVal) Multiply(other Numeric) (Numeric, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isint {
		return IntVal{i.Value * inner_i.Value}, nil
	}
	return FloatVal{float64(i.Value) * inner_f.Value}, nil
}
func (i IntVal) Divide(other Numeric) (Numeric, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isint {
		return IntVal{i.Value / inner_i.Value}, nil
	}
	return FloatVal{float64(i.Value) / inner_f.Value}, nil
}
func (i IntVal) IntDiv(other Numeric) (Numeric, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isint {
		return IntVal{i.Value / inner_i.Value}, nil
	}
	return IntVal{int(float64(i.Value) / inner_f.Value)}, nil
}
func (i IntVal) Modulo(other Numeric) (Numeric, error) {
	inner_i, isint := other.(IntVal)
	_, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isint {
		return IntVal{i.Value % inner_i.Value}, nil
	}
	return nil, fmt.Errorf("Right hand value must be integer")
}
func (i FloatVal) Invert() (Numeric, error) {
	return FloatVal{-i.Value}, nil
}
func (i FloatVal) Subtract(other Numeric) (Numeric, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isfloat {
		return FloatVal{i.Value - inner_f.Value}, nil
	}
	return FloatVal{i.Value - float64(inner_i.Value)}, nil
}
func (i FloatVal) Multiply(other Numeric) (Numeric, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isfloat {
		return FloatVal{i.Value * inner_f.Value}, nil
	}
	return FloatVal{i.Value * float64(inner_i.Value)}, nil
}
func (i FloatVal) Divide(other Numeric) (Numeric, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isfloat {
		return FloatVal{i.Value / inner_f.Value}, nil
	}
	return FloatVal{i.Value / float64(inner_i.Value)}, nil
}
func (i FloatVal) IntDiv(other Numeric) (Numeric, error) {
	inner_i, isint := other.(IntVal)
	inner_f, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isfloat {
		return IntVal{int(i.Value / inner_f.Value)}, nil
	}
	return FloatVal{i.Value - float64(inner_i.Value)}, nil
}
func (i FloatVal) Modulo(other Numeric) (Numeric, error) {
	inner_i, isint := other.(IntVal)
	_, isfloat := other.(FloatVal)
	if !isint && !isfloat {
		panic("Unreachable!")
	}
	if isfloat {
		return nil, fmt.Errorf("Right hand value must be integer")
	}
	whole := int(i.Value) % inner_i.Value
	frac := i.Value - float64(inner_i.Value)
	return FloatVal{float64(whole) + frac}, nil
}

func (i ListVal) Index(idx IntVal) (Val, error) {
	return i.Stored[idx.Value], nil // bounds check here?
}
func (i StrVal) Index(idx IntVal) (Val, error) {
	return CharVal{[]rune(i.Value)[idx.Value]}, nil // bounds check here?
}

func binCompat[T Val](a, b Val, oper string) (T, T, error) {
	ia, aok := a.(T)
	ib, bok := b.(T)
	if !aok || !bok {
		// fmt.Println(a, b)
		return ia, ib, fmt.Errorf("\"%s\" binary operation not defined between %s and %s", oper, a.Name(), b.Name())
	}
	return ia, ib, nil
}
func unCompat[T Val](a Val, oper string) (T, error) {
	ia, aok := a.(T)
	if !aok {
		return ia, fmt.Errorf("\"%s\" unary operation not defined on %s", oper, a.Name())
	}
	return ia, nil
}

func OpEqual(a, b Val, name string) (Logical, error) {
	ia, ib, err := binCompat[Comparable](a, b, name)
	if err != nil {
		return nil, err
	}
	return ia.Equal(ib), nil
}
func OpNotEqual(a, b Val, name string) (Logical, error) {
	ia, ib, err := binCompat[Comparable](a, b, name)
	if err != nil {
		return nil, err
	}
	temp := ia.Equal(ib)
	temp.Not()
	return temp, nil
}

func OpNot(a Val, name string) (Logical, error) {
	ia, err := unCompat[Logical](a, name)
	return ia.Not(), err
}
func OpAnd(a, b Val, name string) (Logical, error) {
	ia, ib, err := binCompat[Logical](a, b, name)
	if err != nil {
		return nil, err
	}
	return ia.And(ib), err
}
func OpOr(a, b Val, name string) (Logical, error) {
	ia, ib, err := binCompat[Logical](a, b, name)
	if err != nil {
		return nil, err
	}
	return ia.Or(ib), err
}
func OpAdd(a, b Val, name string) (Addable, error) {
	ia, ib, err := binCompat[Addable](a, b, name)
	if err != nil {
		return nil, err
	}
	return ia.Add(ib)
}
func OpGreater(a, b Val, name string) (Logical, error) {
	ia, ib, err := binCompat[Ordered](a, b, name)
	if err != nil {
		return nil, err
	}
	return ia.Greater(ib)
}
func OpLess(a, b Val, name string) (Logical, error) {
	ia, ib, err := binCompat[Ordered](a, b, name)
	if err != nil {
		return nil, err
	}
	step1, _ := ia.Greater(ib)
	step2 := ia.Equal(ib)
	step1 = step1.Not()
	step2 = step2.Not()
	return step1.And(step2), nil
}
func OpGrEqual(a, b Val, name string) (Logical, error) {
	ia, ib, err := binCompat[Ordered](a, b, name)
	if err != nil {
		return nil, err
	}
	step1, _ := ia.Greater(ib)
	step2 := ia.Equal(ib)
	return step1.And(step2), nil

}
func OpLeEqual(a, b Val, name string) (Logical, error) {
	ia, ib, err := binCompat[Ordered](a, b, name)
	if err != nil {
		return nil, err
	}
	step1, _ := (ia).Greater(ib)
	return step1.Not(), nil
}

func OpInvert(a Val, name string) (Numeric, error) {
	ia, err := unCompat[Numeric](a, name)
	if err != nil {
		return nil, err
	}
	return ia.Invert()

}
func OpSubtract(a, b Val, name string) (Numeric, error) {
	ia, ib, err := binCompat[Numeric](a, b, name)
	if err != nil {
		return nil, err
	}
	return ia.Subtract(ib)
}
func OpMultiply(a, b Val, name string) (Numeric, error) {
	ia, ib, err := binCompat[Numeric](a, b, name)
	if err != nil {
		return nil, err
	}
	return ia.Multiply(ib)
}
func OpDivide(a, b Val, name string) (Numeric, error) {
	ia, ib, err := binCompat[Numeric](a, b, name)
	if err != nil {
		return nil, err
	}
	return ia.Divide(ib)
}
func OpIntDiv(a, b Val, name string) (Numeric, error) {
	ia, ib, err := binCompat[Numeric](a, b, name)
	if err != nil {
		return nil, err
	}
	return ia.IntDiv(ib)
}
func OpModulo(a, b Val, name string) (Numeric, error) {
	ia, ib, err := binCompat[Numeric](a, b, name)
	if err != nil {
		return nil, err
	}
	return ia.Modulo(ib)
}

func OpIndex(a, b Val) (Val, error) {
	ia, ok := a.(Indexable)
	if !ok {
		return nil, fmt.Errorf("not indexable")
	}
	ib, ok := b.(IntVal)
	if !ok {
		return nil, fmt.Errorf("not an integer")
	}
	return ia.Index(ib)
}
