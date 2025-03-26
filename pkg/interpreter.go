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
	Types  map[string]TokenType
	Parent *Env
}

func (e *Env) Init(par *Env) {
	if e.Values == nil {
		e.Values = make(map[string]any)
	}
	if e.Types == nil {
		e.Types = make(map[string]TokenType)
	}
	e.Parent = par
}

func (e *Env) Define(id string, typ TokenType) {
	e.Values[id] = nil
	e.Types[id] = typ
}

func (e *Env) Put(id string, typ TokenType, val any, line, pos int) {
	c_typ, ok := e.Types[id]
	//the variable exists, check its type
	if ok {
		if typ != c_typ {
			print_error(line, pos, "Variable type does not match assigned type")
			return
		}
		e.Values[id] = val
		// variable does not exist, check parent if exists
	} else {
		if e.Parent == nil {
			print_error(line, pos, "Variable is undeclared")
			return
		}
		e.Parent.Put(id, typ, val, line, pos)

	}
}
func (e *Env) Get(id string, line, pos int) any {
	val, ok := e.Values[id]
	if !ok {
		if e.Parent != nil {
			return e.Parent.Get(id, line, pos)
		} else {
			print_error(line, pos, "Variable is undeclared")
			return nil
		}
	}
	return val
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

func eval_unary(value any, oper Token) any {
	if had_error {
		return nil
	}
	if oper.Token_type == Not {
		switch v := value.(type) {
		case bool:
			return !v
		default:
			print_error(oper.Line, oper.Start, "Boolean expected")
			return nil
		}
	} else if oper.Token_type == Minus {
		switch v := value.(type) {
		case int:
			return -v
		case float64:
			return -v
		default:
			print_error(oper.Line, oper.Start, "Number expected")
			return nil
		}
	} else {
		panic("Error in unary evaluation!")
	}
}

var had_error bool = false

func print_error(line_num, pos int, message string) {
	fmt.Fprintf(os.Stderr, "[ERROR] Line %d Column %d: %s\n", line_num, pos+1, message)
	had_error = true
}

func got_nul(ltype, rtype string, line, pos int) bool {
	nul := false
	if ltype == "nul" {
		print_error(line, pos, "Left value in binary operator is SEMMI")
		nul = true
	}
	if rtype == "nul" {
		print_error(line, pos, "Right value in binary operator is SEMMI")
		nul = true
	}
	return nul
}

// This is fucking stupid
// I tried to refactor this, but since I can't create generic function literals to pass as arguments
// and I can't create working sum types this "any" shenanigans and this ugly ass 300 line long function must stay
func eval_binary(lval, rval any, oper Token) any {
	if had_error {
		return nil
	}
	ltype, rtype := unwrap_type(lval, rval)
	switch oper.Token_type {
	case And:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		if ltype != "bool" || rtype != "bool" {
			print_error(oper.Line, oper.Start, "Boolean expected")
			return nil
		}
		return lval.(bool) && rval.(bool)
	case Or:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		if ltype != "bool" || rtype != "bool" {
			print_error(oper.Line, oper.Start, "Boolean expected")
			return nil
		}
		return lval.(bool) || rval.(bool)
	case E_equal:
		if ltype != rtype {
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		return lval == rval
	case N_equal:
		if ltype != rtype {
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		return lval != rval
	case Greater:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		unsupp := []string{"bool", "str"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			print_error(oper.Line, oper.Start, "This operator does not support booleans or strings")
			return nil
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) > rval.(float64)
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) > float64(rval.(int))
			}
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		if ltype == "int" {
			return lval.(int) > rval.(int)
		}
		if ltype == "float" {
			return lval.(float64) > rval.(float64)
		}
		if ltype == "char" {
			return lval.(rune) > rval.(rune)
		}
		print_error(oper.Line, oper.Start, "Something went wrong in type checking")
		return nil
	case G_equal:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		unsupp := []string{"bool", "str"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			print_error(oper.Line, oper.Start, "This operator does not support booleans or strings")
			return nil
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) >= rval.(float64)
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) >= float64(rval.(int))
			}
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		if ltype == "int" {
			return lval.(int) >= rval.(int)
		}
		if ltype == "float" {
			return lval.(float64) >= rval.(float64)
		}
		if ltype == "char" {
			return lval.(rune) >= rval.(rune)
		}
		print_error(oper.Line, oper.Start, "Something went wrong in type checking")
		return nil
	case Less:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		unsupp := []string{"bool", "str"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			print_error(oper.Line, oper.Start, "This operator does not support booleans or strings")
			return nil
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) < rval.(float64)
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) < float64(rval.(int))
			}
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		if ltype == "int" {
			return lval.(int) < rval.(int)
		}
		if ltype == "float" {
			return lval.(float64) < rval.(float64)
		}
		if ltype == "char" {
			return lval.(rune) < rval.(rune)
		}
		print_error(oper.Line, oper.Start, "Something went wrong in type checking")
		return nil
	case L_equal:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		unsupp := []string{"bool", "str"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			print_error(oper.Line, oper.Start, "This operator does not support booleans or strings")
			return nil
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) <= rval.(float64)
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) <= float64(rval.(int))
			}
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		if ltype == "int" {
			return lval.(int) <= rval.(int)
		}
		if ltype == "float" {
			return lval.(float64) <= rval.(float64)
		}
		if ltype == "char" {
			return lval.(rune) <= rval.(rune)
		}
		print_error(oper.Line, oper.Start, "Something went wrong in type checking")
		return nil
	case Plus:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		unsupp := []string{"bool", "str", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			print_error(oper.Line, oper.Start, "This operator does not support booleans, strings or characters")
			return nil
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) + rval.(float64)
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) + float64(rval.(int))
			}
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		if ltype == "int" {
			return lval.(int) + rval.(int)
		}
		if ltype == "float" {
			return lval.(float64) + rval.(float64)
		}
		print_error(oper.Line, oper.Start, "Something went wrong in type checking")
		return nil
	case Minus:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		unsupp := []string{"bool", "str", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			print_error(oper.Line, oper.Start, "This operator does not support booleans, strings or characters")
			return nil
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) - rval.(float64)
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) - float64(rval.(int))
			}
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		if ltype == "int" {
			return lval.(int) - rval.(int)
		}
		if ltype == "float" {
			return lval.(float64) - rval.(float64)
		}
		print_error(oper.Line, oper.Start, "Something went wrong in type checking")
		return nil
	case Star:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		unsupp := []string{"bool", "str", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			print_error(oper.Line, oper.Start, "This operator does not support booleans, strings or characters")
			return nil
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) * rval.(float64)
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) * float64(rval.(int))
			}
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		if ltype == "int" {
			return lval.(int) * rval.(int)
		}
		if ltype == "float" {
			return lval.(float64) * rval.(float64)
		}
		print_error(oper.Line, oper.Start, "Something went wrong in type checking")
		return nil
	case Slash:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		unsupp := []string{"bool", "str", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			print_error(oper.Line, oper.Start, "This operator does not support booleans, strings or characters")
			return nil
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return float64(lval.(int)) / rval.(float64)
			}
			if ltype == "float" && rtype == "int" {
				return lval.(float64) / float64(rval.(int))
			}
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		if ltype == "int" {
			return float64(lval.(int)) / float64(rval.(int))
		}
		if ltype == "float" {
			return lval.(float64) / rval.(float64)
		}
		print_error(oper.Line, oper.Start, "Something went wrong in type checking")
		return nil
	case Sl_slash:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		unsupp := []string{"bool", "str", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			print_error(oper.Line, oper.Start, "This operator does not support booleans, strings or characters")
			return nil
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				return lval.(int) / int(rval.(float64))
			}
			if ltype == "float" && rtype == "int" {
				return int(lval.(float64)) / rval.(int)
			}
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		if ltype == "int" {
			return lval.(int) / rval.(int)
		}
		if ltype == "float" {
			return int(lval.(float64) / rval.(float64))
		}
		print_error(oper.Line, oper.Start, "Something went wrong in type checking")
		return nil
	case Percent:
		if got_nul(ltype, rtype, oper.Line, oper.Start) {
			return nil
		}
		unsupp := []string{"bool", "str", "char"}
		if slices.Contains(unsupp, ltype) || slices.Contains(unsupp, rtype) {
			print_error(oper.Line, oper.Start, "This operator does not support booleans, strings or characters")
			return nil
		}
		if ltype != rtype {
			//convert type
			if ltype == "int" && rtype == "float" {
				print_error(oper.Line, oper.Start, "Right hand value must be integer")
				return nil
			}
			if ltype == "float" && rtype == "int" {
				whole := int(lval.(float64)) % rval.(int)
				frac := lval.(float64) - float64(int(lval.(float64)))
				if frac == 0 {
					return whole
				}
				return float64(whole) + frac
			}
			print_error(oper.Line, oper.Start, "Mismatched types not supported")
			return nil
		}
		if ltype == "int" {
			return lval.(int) % rval.(int)
		}
		if ltype == "float" {
			print_error(oper.Line, oper.Start, "Right hand value must be integer")
			return nil
		}
		print_error(oper.Line, oper.Start, "Something went wrong in type checking")
		return nil
	default:
		panic("Error in binary evaluation!")
	}
}
func (node *TreeNode) eval(env *Env) any {
	node_type := 0
	if node.left != nil {
		node_type += 1
	}
	if node.right != nil {
		node_type += 10
	}

	switch node_type {
	//Literal or variable
	case 0:
		if node.oper.Token_type == Identifier {
			//variable
			return env.Get(node.oper.Lexeme, node.oper.Line, node.oper.Start)
		} else {
			//literal
			return node.oper.Literal
		}
	//Group
	case 1:
		return node.left.eval(env)
	//Unary
	case 10:
		return eval_unary(node.right.eval(env), node.oper)
	//Binary
	case 11:
		return eval_binary(node.left.eval(env), node.right.eval(env), node.oper)
	default:
		panic("Error in AST node_type")
	}
}

func (node *TreeNode) print() string {
	ret := fmt.Sprintf("%v", node.oper.Lexeme) + " "
	if node.left != nil && node.right != nil {
		ret += "(" + node.left.print() + node.right.print() + ")"
	} else if node.left != nil {
		ret = "(group " + node.left.print() + ")"
	} else if node.right != nil {
		ret += "(" + node.right.print() + ")"
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

func (cmd *Command) Interpret(env *Env) {
	switch cmd.stmt_type {
	case expr_cmd:
		cmd.head.eval(env)
	case print_cmd:
		out := cmd.head.eval(env)
		if had_error {
			return
		}
		fmt.Println(stringify(out))
	case read_cmd:
		if cmd.head.oper.Token_type != Identifier {
			print_error(cmd.head.oper.Line, cmd.head.oper.Start, "Reading can only be into variables")
			return
		}
		reader := bufio.NewReader(os.Stdin)

		fmt.Printf("%s ?>", cmd.head.oper.Lexeme)
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		//probably convert type here?
		num, err := strconv.Atoi(text)
		if err != nil {
			panic(err)
		}
		env.Put(cmd.head.oper.Lexeme, Type_num, num, cmd.head.oper.Line, cmd.head.oper.Start)
	case vardef_cmd:
		env.Define(cmd.id, cmd.vtype)
		fallthrough
	case assign_cmd:
		if cmd.head != nil {
			val := cmd.head.eval(env)
			if had_error {
				return
			}
			var etype TokenType
			switch val.(type) {
			case bool:
				etype = Type_bool
			case rune:
				etype = Type_char
			case string:
				etype = Type_str
			case int:
				etype = Type_num
			case float64:
				etype = Type_num
			default:
				etype = Nul
			}
			env.Put(cmd.id, etype, val, cmd.head.oper.Line, 0)
		}
	case prog_cmd:
		g_env := Env{}
		g_env.Init(nil)
		for _, b_cmd := range cmd.body {
			b_cmd.Interpret(&g_env)
		}
	// other commands are always inside the prog body
	// meaning we create a new env and link it
	// to the env we got as an argument
	case if_cmd:
		decide := cmd.head.eval(env)
		switch decide.(type) {
		case bool:
			break
		default:
			print_error(cmd.head.oper.Line, cmd.head.oper.Start, "Type must be boolean expression")
		}
		if had_error {
			return
		}
		if decide.(bool) {
			cmd.body[0].Interpret(env)
		} else {
			if cmd.body[1] == nil {
				return
			}
			cmd.body[1].Interpret(env)
		}
	case for_cmd:
		rep := cmd.head.eval(env)
		switch rep.(type) {
		case int:
			break
		default:
			print_error(cmd.head.oper.Line, cmd.head.oper.Start, "Type must be integer expression")
		}
		if had_error {
			return
		}
		if cmd.body[1] == nil {
			return
		}

		l_env := Env{}
		l_env.Init(env)
		idx_name := ""
		if cmd.body[0] != nil {
			cmd.body[0].Interpret(&l_env)
			idx_name = cmd.body[0].id
		}

		for_block := cmd.body[1]
		for range rep.(int) {
			for_block.Interpret(&l_env)
			if idx_name != "" {
				l_env.Values[idx_name] = l_env.Values[idx_name].(int) + 1
			}
		}
	case while_cmd:
		cond := cmd.head.eval(env)
		switch cond.(type) {
		case bool:
			break
		default:
			print_error(cmd.head.oper.Line, cmd.head.oper.Start, "Type must be boolean expression")
		}
		if had_error {
			return
		}
		if cmd.body[1] == nil {
			return
		}

		l_env := Env{}
		l_env.Init(env)
		idx_name := ""
		if cmd.body[0] != nil {
			cmd.body[0].Interpret(&l_env)
			idx_name = cmd.body[0].id
		}

		for_block := cmd.body[1]
		for cond.(bool) {
			for_block.Interpret(&l_env)
			if idx_name != "" {
				l_env.Values[idx_name] = l_env.Values[idx_name].(int) + 1
			}
			cond = cmd.head.eval(env)
		}
	case block:
		l_env := Env{}
		l_env.Init(env)
		for _, c := range cmd.body {
			c.Interpret(&l_env)
		}
	}
}
func (cmd *Command) Print(in string) {
	// fmt.Print(in)
	fmt.Printf("%s(%v ", in, cmd.stmt_type)
	if cmd.stmt_type == vardef_cmd {
		fmt.Printf("%s, %s", cmd.id, cmd.vtype)
	}
	if cmd.head != nil {
		fmt.Printf(" Head: %s", cmd.head.print())
	}
	if len(cmd.body) > 0 {
		fmt.Print(" Body:\n")
		for _, b_cmd := range cmd.body {
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
