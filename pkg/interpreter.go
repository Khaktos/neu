package neu

import (
	"bufio"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Env struct {
	Values map[string]any
	Types  map[string]ValueType
	Parent *Env
}

func (e *Env) Init(par *Env) {
	if e.Values == nil {
		e.Values = make(map[string]any)
	}
	if e.Types == nil {
		e.Types = make(map[string]ValueType)
	}
	e.Parent = par
}

func (e *Env) Define(id string, typ ValueType) {
	e.Values[id] = nil
	e.Types[id] = typ
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

func (e *Env) Put(id string, typ ValueType, val any, line, pos int) error {
	c_typ, ok := e.Types[id]
	//the variable exists, check its type
	if ok {
		if !isSameType(typ, c_typ) {
			return &RunError{line, pos, "Variable type does not match assigned type"}
		}
		e.Values[id] = val
		// variable does not exist, check parent if exists
	} else {
		if e.Parent == nil {
			return &RunError{line, pos, "Variable is undeclared"}
		}
		err := e.Parent.Put(id, typ, val, line, pos)
		if err != nil {
			return err
		}

	}
	return nil
}
func (e *Env) Get(id string, line, pos int) (any, error) {
	val, ok := e.Values[id]
	if !ok {
		if e.Parent != nil {
			return e.Parent.Get(id, line, pos)
		} else {
			return nil, &RunError{line, pos, "Variable is undeclared"}
		}
	}
	return val, nil
}
func (e *Env) Get_type(id string, line, pos int) (ValueType, error) {
	c_typ, ok := e.Types[id]
	if !ok {
		return ValueType{}, &RunError{line, pos, "Variable is undeclared"}
	}
	return c_typ, nil
}

type RunError struct {
	line int
	pos  int
	msg  string
}

func (e *RunError) Error() string {
	return fmt.Sprintf("[ERROR] Runtime: Line %d Column %d: %s", e.line, e.pos+1, e.msg)
}

func unwrap_type(left, right any) (ltype, rtype string) {
	switch left.(type) {
	case bool:
		ltype = "bool"
	case rune:
		ltype = "char"
	case string:
		ltype = "str"
	case int:
		ltype = "int"
	case float64:
		ltype = "float"
	default:
		ltype = "nul"
	}
	switch right.(type) {
	case bool:
		rtype = "bool"
	case rune:
		rtype = "char"
	case string:
		rtype = "str"
	case int:
		rtype = "int"
	case float64:
		rtype = "float"
	default:
		rtype = "nul"
	}
	return
}

func eval_unary(value any, oper Token) (any, error) {
	switch oper.Token_type {
	case Not:
		switch v := value.(type) {
		case bool:
			return !v, nil
		default:
			return nil, &RunError{oper.Line, oper.Start, "Boolean expected"}
		}
	case Minus:
		switch v := value.(type) {
		case int:
			return -v, nil
		case float64:
			return -v, nil
		default:
			return nil, &RunError{oper.Line, oper.Start, "Number expected"}
		}
	default:
		panic("Error in unary evaluation!")
	}
}

func got_nul(ltype, rtype string, line, pos int) error {
	if ltype == "nul" {
		return &RunError{line, pos, "Left value in binary operator is SEMMI"}
	}
	if rtype == "nul" {
		return &RunError{line, pos, "Right value in binary operator is SEMMI"}
	}
	return nil
}

// This is fucking stupid
// I tried to refactor this, but since I can't create generic function literals to pass as arguments
// and I can't create working sum types this "any" shenanigans and this ugly ass 300 line long function must stay
func eval_binary(lval, rval any, oper Token) (any, error) {
	ltype, rtype := unwrap_type(lval, rval)
	switch oper.Token_type {
	case And:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		if ltype != "bool" || rtype != "bool" {
			return nil, &RunError{oper.Line, oper.Start, "Boolean expected"}
		}
		return lval.(bool) && rval.(bool), nil
	case Or:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		if ltype != "bool" || rtype != "bool" {
			return nil, &RunError{oper.Line, oper.Start, "Boolean expected"}
		}
		return lval.(bool) || rval.(bool), nil
	case E_equal:
		// if ltype != rtype {
		// 	return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		// }
		return lval == rval, nil
	case N_equal:
		// if ltype != rtype {
		// 	return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		// }
		return lval != rval, nil
	case Greater:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		unsupp := []string{"bool", "str"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			return nil, &RunError{oper.Line, oper.Start, "This operator does not support booleans or strings"}
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) > rval.(float64), nil
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) > float64(rval.(int)), nil
			}
			return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		}
		if ltype == "int" {
			return lval.(int) > rval.(int), nil
		}
		if ltype == "float" {
			return lval.(float64) > rval.(float64), nil
		}
		if ltype == "char" {
			return lval.(rune) > rval.(rune), nil
		}
		return nil, &RunError{oper.Line, oper.Start, "Something went wrong in type checking"}
	case G_equal:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		unsupp := []string{"bool", "str"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			return nil, &RunError{oper.Line, oper.Start, "This operator does not support booleans or strings"}
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) >= rval.(float64), nil
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) >= float64(rval.(int)), nil
			}
			return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		}
		if ltype == "int" {
			return lval.(int) >= rval.(int), nil
		}
		if ltype == "float" {
			return lval.(float64) >= rval.(float64), nil
		}
		if ltype == "char" {
			return lval.(rune) >= rval.(rune), nil
		}
		return nil, &RunError{oper.Line, oper.Start, "Something went wrong in type checking"}
	case Less:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		unsupp := []string{"bool", "str"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			return nil, &RunError{oper.Line, oper.Start, "This operator does not support booleans or strings"}
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) < rval.(float64), nil
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) < float64(rval.(int)), nil
			}
			return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		}
		if ltype == "int" {
			return lval.(int) < rval.(int), nil
		}
		if ltype == "float" {
			return lval.(float64) < rval.(float64), nil
		}
		if ltype == "char" {
			return lval.(rune) < rval.(rune), nil
		}
		return nil, &RunError{oper.Line, oper.Start, "Something went wrong in type checking"}
	case L_equal:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		unsupp := []string{"bool", "str"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			return nil, &RunError{oper.Line, oper.Start, "This operator does not support booleans or strings"}
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) <= rval.(float64), nil
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) <= float64(rval.(int)), nil
			}
			return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		}
		if ltype == "int" {
			return lval.(int) <= rval.(int), nil
		}
		if ltype == "float" {
			return lval.(float64) <= rval.(float64), nil
		}
		if ltype == "char" {
			return lval.(rune) <= rval.(rune), nil
		}
		return nil, &RunError{oper.Line, oper.Start, "Something went wrong in type checking"}
	case Plus:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		unsupp := []string{"bool", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			return nil, &RunError{oper.Line, oper.Start, "This operator does not support booleans or characters"}
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) + rval.(float64), nil
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) + float64(rval.(int)), nil
			}
			return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		}
		if ltype == "int" {
			return lval.(int) + rval.(int), nil
		}
		if ltype == "float" {
			return lval.(float64) + rval.(float64), nil
		}
		if ltype == "str" {
			return fmt.Sprintf("%s%s", lval.(string), rval.(string)), nil
		}
		return nil, &RunError{oper.Line, oper.Start, "Something went wrong in type checking"}
	case Minus:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		unsupp := []string{"bool", "str", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			return nil, &RunError{oper.Line, oper.Start, "This operator does not support booleans, strings or characters"}
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) - rval.(float64), nil
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) - float64(rval.(int)), nil
			}
			return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		}
		if ltype == "int" {
			return lval.(int) - rval.(int), nil
		}
		if ltype == "float" {
			return lval.(float64) - rval.(float64), nil
		}
		return nil, &RunError{oper.Line, oper.Start, "Something went wrong in type checking"}
	case Star:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		unsupp := []string{"bool", "str", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			return nil, &RunError{oper.Line, oper.Start, "This operator does not support booleans, strings or characters"}
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) * rval.(float64), nil
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) * float64(rval.(int)), nil
			}
			return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		}
		if ltype == "int" {
			return lval.(int) * rval.(int), nil
		}
		if ltype == "float" {
			return lval.(float64) * rval.(float64), nil
		}
		return nil, &RunError{oper.Line, oper.Start, "Something went wrong in type checking"}
	case Slash:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		unsupp := []string{"bool", "str", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			return nil, &RunError{oper.Line, oper.Start, "This operator does not support booleans, strings or characters"}
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) / rval.(float64), nil
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) / float64(rval.(int)), nil
			}
			return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		}
		if ltype == "int" {
			return float64(lval.(int)) / float64(rval.(int)), nil
		}
		if ltype == "float" {
			return lval.(float64) / rval.(float64), nil
		}
		return nil, &RunError{oper.Line, oper.Start, "Something went wrong in type checking"}
	case Sl_slash:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		unsupp := []string{"bool", "str", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			return nil, &RunError{oper.Line, oper.Start, "This operator does not support booleans, strings or characters"}
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return lval.(int) / int(rval.(float64)), nil
			}
			if ltype == "float" && rtype == "int" {
				return int(lval.(float64)) / rval.(int), nil
			}
			return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		}
		if ltype == "int" {
			return lval.(int) / rval.(int), nil
		}
		if ltype == "float" {
			return int(lval.(float64) / rval.(float64)), nil
		}
		return nil, &RunError{oper.Line, oper.Start, "Something went wrong in type checking"}
	case Percent:
		err := got_nul(ltype, rtype, oper.Line, oper.Start)
		if err != nil {
			return nil, err
		}
		unsupp := []string{"bool", "str", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			return nil, &RunError{oper.Line, oper.Start, "This operator does not support booleans, strings or characters"}
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return nil, &RunError{oper.Line, oper.Start, "Right hand value must be integer"}
			}
			if ltype == "float" && rtype == "int" {
				whole := int(lval.(float64)) % rval.(int)
				frac := lval.(float64) - float64(int(lval.(float64)))
				if frac == 0 {
					return whole, nil
				}
				return float64(whole) + frac, nil
			}
			return nil, &RunError{oper.Line, oper.Start, "Mismatched types not supported"}
		}
		if ltype == "int" {
			return lval.(int) % rval.(int), nil
		}
		if ltype == "float" {
			return nil, &RunError{oper.Line, oper.Start, "Right hand value must be integer"}
		}
		return nil, &RunError{oper.Line, oper.Start, "Something went wrong in type checking"}
	case L_brace:
		if rtype == "nul" {
			return nil, &RunError{oper.Line, oper.Start, "Index cannot be SEMMI"}
		}
		fmt.Println(lval)
		return nil, &RunError{oper.Line, oper.Start, "INDEXING"}
	default:
		panic("Error in binary evaluation!")
	}
}
func (node *TreeNode) eval(env *Env) (any, error) {
	node_type := 0
	if node.Left != nil {
		node_type += 1
	}
	if node.Right != nil {
		node_type += 10
	}

	switch node_type {
	//Literal or variable
	case 0:
		if node.Oper.Token_type == Identifier {
			//variable
			ret, err := env.Get(node.Oper.Lexeme, node.Oper.Line, node.Oper.Start)
			if err != nil {
				return nil, err
			}
			return ret, nil
		} else {
			//literal
			return node.Oper.Literal, nil
		}
	//Group
	case 1:
		return node.Left.eval(env)
	//Unary
	case 10:
		val, err := node.Right.eval(env)
		if err != nil {
			return nil, err
		}
		return eval_unary(val, node.Oper)
	//Binary
	case 11:
		lval, lerr := node.Left.eval(env)
		if lerr != nil {
			return nil, lerr
		}
		rval, rerr := node.Right.eval(env)
		if rerr != nil {
			return nil, rerr
		}
		return eval_binary(lval, rval, node.Oper)
	default:
		panic("Error in AST node_type")
	}
}

func (node *TreeNode) print() string {
	ret := fmt.Sprintf("%v", node.Oper.Lexeme) + " "
	if node.Left != nil && node.Right != nil {
		ret += "(" + node.Left.print() + node.Right.print() + ")"
	} else if node.Left != nil {
		ret = "(group " + node.Left.print() + ")"
	} else if node.Right != nil {
		ret += "(" + node.Right.print() + ")"
	}
	return ret
}
func stringify(stuf any) string {
	switch val := stuf.(type) {
	case float64:
		return fmt.Sprintf("%f", val)
	case int:
		return fmt.Sprintf("%d", val)
	case bool:
		if val {
			return "IGAZ"
		}
		return "HAMIS"
	case rune:
		return fmt.Sprintf("%c", val)
	case string:
		return val
	case nil:
		return "SEMMI"
	default:
		return "Unknown type"
	}
}

var reader *bufio.Reader

func do_read(env *Env, cmd *Command) error {
	if cmd.Head.Oper.Token_type != Identifier {
		return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, "Reading can only be into variables"}
	}
	if reader == nil {
		reader = bufio.NewReader(os.Stdin)
	}

	fmt.Printf("%s ?>", cmd.Head.Oper.Lexeme)
	valtype, err := env.Get_type(cmd.Head.Oper.Lexeme, cmd.Head.Oper.Line, cmd.Head.Oper.Start)
	if err != nil {
		return err
	}
	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return nil
	}
	var val any

	if isSameType(valtype, BaseNum) {
		val, err = strconv.Atoi(text)
		if err != nil {
			val, err = strconv.ParseFloat(text, 64)
			if err != nil {
				return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, "Can not convert to NUM type"}
			}
		}
	}
	if isSameType(valtype, BaseChar) {
		if len(text) > 1 || len(text) == 0 {
			return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, "Can not convert to KAR type"}
		}
		val = []rune(text)[0]
	}
	if isSameType(valtype, BaseBool) {
		if text != "IGAZ" && text != "HAMIS" {
			return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, "Can not convert to LOG type"}
		}
		val = text == "IGAZ"
	}
	if isSameType(valtype, BaseStr) {
		val = text
	}

	err = env.Put(cmd.Head.Oper.Lexeme, valtype, val, cmd.Head.Oper.Line, cmd.Head.Oper.Start)
	if err != nil {
		return err
	}
	return nil
}

func (cmd *Command) Interpret(env *Env) error {
	switch cmd.Stmt_type {
	case Expr_cmd:
		cmd.Head.eval(env)
	case Print_cmd:
		out, err := cmd.Head.eval(env)
		if err != nil {
			return err
		}
		fmt.Print(stringify(out))
	case Read_cmd:
		err := do_read(env, cmd)
		if err != nil {
			return err
		}
	case Vardef_cmd:
		env.Define(cmd.Def.Id, cmd.Def.Valtype)
		fallthrough
	case Assign_cmd:
		if cmd.Head != nil {
			val, err := cmd.Head.eval(env)
			if err != nil {
				return err
			}
			var etype ValueType
			switch val.(type) {
			case bool:
				etype = BaseBool
			case rune:
				etype = BaseChar
			case string:
				etype = BaseStr
			case int:
				etype = BaseNum
			case float64:
				etype = BaseNum
			default:
				etype = ValueType{}
			}
			if len(cmd.Body) > 0 {
				for _, c := range cmd.Body {
					ix, err := c.Head.eval(env)
					if err != nil {
						return err
					}
					idx, ok := ix.(int)
					if !ok {
						return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, "Indexing variable must be integer"}
					}
					pv, err := env.Get(cmd.Def.Id, cmd.Head.Oper.Line, cmd.Head.Oper.Start)
					if err != nil {
						return err
					}
					prev, ok := pv.([]any)
					if !ok {
						panic("Something went wrong in list element assignment")
					}
					prev[idx] = val
					env.Put(cmd.Def.Id, cmd.Def.Valtype, prev, cmd.Head.Oper.Line, cmd.Head.Oper.Start)
				}
			} else {
				err = env.Put(cmd.Def.Id, etype, val, cmd.Head.Oper.Line, 0)
			}
			if err != nil {
				return err
			}
		}
	case Prog_cmd:
		g_env := Env{}
		g_env.Init(nil)
		for _, b_cmd := range cmd.Body {
			err := b_cmd.Interpret(&g_env)
			if err != nil {
				return err
			}
		}
	// other commands are always inside the prog body
	// meaning we create a new env and link it
	// to the env we got as an argument
	case If_cmd:
		decide, err := cmd.Head.eval(env)
		if err != nil {
			return err
		}
		switch decide.(type) {
		case bool:
			break
		default:
			msg := fmt.Sprintf("Value of expression: %s is not a boolean", stringify(decide))
			return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, msg}
		}
		if decide.(bool) {
			err := cmd.Body[0].Interpret(env)
			if err != nil {
				return err
			}
		} else {
			if cmd.Body[1] == nil {
				return nil
			}
			err := cmd.Body[1].Interpret(env)
			if err != nil {
				return err
			}
		}
	case For_cmd:
		rep, err := cmd.Head.eval(env)
		if err != nil {
			return err
		}
		switch rep.(type) {
		case int:
			break
		default:
			msg := fmt.Sprintf("Value of expression: %s is not an integer", stringify(rep))
			return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, msg}
		}
		if cmd.Body[1] == nil {
			return nil
		}

		l_env := Env{}
		l_env.Init(env)
		idx_name := ""
		if cmd.Body[0] != nil {
			err := cmd.Body[0].Interpret(&l_env)
			if err != nil {
				return err
			}
			idx_name = cmd.Body[0].Def.Id
		}

		for_block := cmd.Body[1]
		for range rep.(int) {
			err := for_block.Interpret(&l_env)
			if err != nil {
				return err
			}
			if idx_name != "" {
				l_env.Values[idx_name] = l_env.Values[idx_name].(int) + 1
			}
		}
	case While_cmd:
		cond, err := cmd.Head.eval(env)
		if err != nil {
			return err
		}
		switch cond.(type) {
		case bool:
			break
		default:
			msg := fmt.Sprintf("Value of expression: %s is not a boolean", stringify(cond))
			return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, msg}
		}
		if cmd.Body[1] == nil {
			return nil
		}

		l_env := Env{}
		l_env.Init(env)
		idx_name := ""
		if cmd.Body[0] != nil {
			err := cmd.Body[0].Interpret(&l_env)
			if err != nil {
				return err
			}
			idx_name = cmd.Body[0].Def.Id
		}

		for_block := cmd.Body[1]
		for cond.(bool) {
			err := for_block.Interpret(&l_env)
			if err != nil {
				return err
			}
			if idx_name != "" {
				l_env.Values[idx_name] = l_env.Values[idx_name].(int) + 1
			}
			cond, err = cmd.Head.eval(env)
			if err != nil {
				return err
			}
		}
	case Block:
		l_env := Env{}
		l_env.Init(env)
		for _, c := range cmd.Body {
			err := c.Interpret(&l_env)
			if err != nil {
				return err
			}
		}
	case Listdef_cmd:
		s, err := cmd.Head.eval(env)
		if err != nil {
			return err
		}
		size, ok := s.(int)
		if !ok {
			return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, "List size is not an integer"}
		}
		list := make([]any, size)
		env.Define(cmd.Def.Id, cmd.Def.Valtype)
		env.Put(cmd.Def.Id, cmd.Def.Valtype, list, cmd.Head.Oper.Line, 0)
	default:
		fmt.Println("New command with undefined behavior:")
		cmd.Print("  >  ")
	}
	return nil
}
func (cmd *Command) Print(in string) {
	fmt.Printf("%s(%v ", in, cmd.Stmt_type)
	if cmd.Stmt_type == Vardef_cmd || cmd.Stmt_type == Listdef_cmd {
		fmt.Printf("%s, %s", cmd.Def.Id, cmd.Def.Valtype)
	}
	if cmd.Stmt_type == Assign_cmd {
		fmt.Printf("%s,", cmd.Def.Id)
	}
	if cmd.Head != nil {
		fmt.Printf(" Head: %s", cmd.Head.print())
	}
	if len(cmd.Body) > 0 {
		fmt.Print(" Body:\n")
		for _, b_cmd := range cmd.Body {
			if b_cmd == nil {
				continue
			}
			b_cmd.Print(fmt.Sprintf("  %s", in))
		}
	} else {
		in = ""
	}
	fmt.Printf("%s)\n", in)
}
