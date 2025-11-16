package neu

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Env struct {
	Values map[string]Val
	Types  map[string]ValueType
	Parent *Env
}

func (e *Env) Init(par *Env) {
	if e.Values == nil {
		e.Values = make(map[string]Val)
	}
	if e.Types == nil {
		e.Types = make(map[string]ValueType)
	}
	e.Parent = par
}

func (e *Env) Define(id string, typ ValueType) {
	e.Values[id] = &NullVal{}
	e.Types[id] = typ
}

func (e *Env) Put(id string, typ ValueType, val Val, line, pos int) error {
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
func (e *Env) Get(id string, line, pos int) (Val, error) {
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

func eval_unary(value Val, oper Token) (Val, error) {
	switch oper.Token_type {
	case Not:
		return OpNot(value, oper.Lexeme)
	case Minus:
		return OpInvert(value, oper.Lexeme)
	default:
		panic("Error in unary evaluation!")
	}
}

func eval_binary(lval, rval Val, oper Token) (Val, error) {
	switch oper.Token_type {
	case And:
		return OpAnd(lval, rval, oper.Lexeme)
	case Or:
		return OpOr(lval, rval, oper.Lexeme)
	case E_equal:
		return OpEqual(lval, rval, oper.Lexeme)
	case N_equal:
		return OpNotEqual(lval, rval, oper.Lexeme)
	case Greater:
		return OpGreater(lval, rval, oper.Lexeme)
	case G_equal:
		return OpGrEqual(lval, rval, oper.Lexeme)
	case Less:
		return OpLess(lval, rval, oper.Lexeme)
	case L_equal:
		return OpLeEqual(lval, rval, oper.Lexeme)
	case Plus:
		return OpAdd(lval, rval, oper.Lexeme)
	case Minus:
		return OpSubtract(lval, rval, oper.Lexeme)
	case Star:
		return OpMultiply(lval, rval, oper.Lexeme)
	case Slash:
		return OpDivide(lval, rval, oper.Lexeme)
	case Sl_slash:
		return OpIntDiv(lval, rval, oper.Lexeme)
	case Percent:
		return OpModulo(lval, rval, oper.Lexeme)
	case L_brace:
		return OpIndex(lval, rval)
		// panic("INDEXING")
		// if rtype == "nul" {
		// 	return nil, &RunError{oper.Line, oper.Start, "Index cannot be SEMMI"}
		// }
		// fmt.Println(lval)
		// return nil, &RunError{oper.Line, oper.Start, "INDEXING"}
	default:
		panic("Error in binary evaluation!")
	}
}
func (node *TreeNode) eval(env *Env) (Val, error) {
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
		ret, err := node.Left.eval(env)
		return ret, err
	//Unary
	case 10:
		val, err := node.Right.eval(env)
		if err != nil {
			return nil, err
		}
		ret, err := eval_unary(val, node.Oper)
		if err != nil {
			err = &RunError{node.Oper.Line, node.Oper.Start, err.Error()}
		}
		return ret, err
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
		ret, err := eval_binary(lval, rval, node.Oper)
		if err != nil {
			err = &RunError{node.Oper.Line, node.Oper.Start, err.Error()}
		}
		return ret, err
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
func stringify(stuf Val) string {
	switch val := stuf.(type) {
	case *FloatVal:
		return fmt.Sprintf("%f", val.Value)
	case *IntVal:
		return fmt.Sprintf("%d", val.Value)
	case *BoolVal:
		if val.Value {
			return "IGAZ"
		}
		return "HAMIS"
	case *CharVal:
		return fmt.Sprintf("%c", val.Value)
	case *StrVal:
		return val.Value
	case *NullVal:
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

	var val Val

	if isSameType(valtype, BaseNum) {
		temp, err := strconv.Atoi(text)
		if err != nil {
			temp, err := strconv.ParseFloat(text, 64)
			if err != nil {
				return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, "Can not convert to NUM type"}
			}
			val = &FloatVal{temp}
		} else {
			val = &IntVal{temp}
		}
	}
	if isSameType(valtype, BaseChar) {
		if len(text) > 1 || len(text) == 0 {
			return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, "Can not convert to KAR type"}
		}
		val = &CharVal{[]rune(text)[0]}
	}
	if isSameType(valtype, BaseBool) {
		if text != "IGAZ" && text != "HAMIS" {
			return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, "Can not convert to LOG type"}
		}
		val = &BoolVal{text == "IGAZ"}
	}
	if isSameType(valtype, BaseStr) {
		val = &StrVal{text}
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
			case *BoolVal:
				etype = BaseBool
			case *CharVal:
				etype = BaseChar
			case *StrVal:
				etype = BaseStr
			case *IntVal:
				etype = BaseNum
			case *FloatVal:
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
					idx, ok := ix.Get().(int)
					if !ok {
						return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, "Indexing variable must be integer"}
					}
					vlist, err := env.Get(cmd.Def.Id, cmd.Head.Oper.Line, cmd.Head.Oper.Start)
					if err != nil {
						return err
					}
					list, ok := vlist.(*ListVal)
					if !ok {
						panic("Something went wrong in list element assignment")
					}
					list.Stored[idx] = val
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
		case *BoolVal:
			break
		default:
			msg := fmt.Sprintf("Value of expression: %s is not a boolean", stringify(decide))
			return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, msg}
		}
		if decide.(*BoolVal).Get().(bool) {
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
		case *IntVal:
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
		for range rep.Get().(int) {
			err := for_block.Interpret(&l_env)
			if err != nil {
				return err
			}
			if idx_name != "" {
				l_env.Values[idx_name] = &IntVal{l_env.Values[idx_name].(*IntVal).Value + 1}
			}
		}
	case While_cmd:
		cond, err := cmd.Head.eval(env)
		if err != nil {
			return err
		}
		switch cond.(type) {
		case *BoolVal:
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
		for cond.Get().(bool) {
			err := for_block.Interpret(&l_env)
			if err != nil {
				return err
			}
			if idx_name != "" {
				temp := l_env.Values[idx_name].(*IntVal)
				temp.Value = temp.Value + 1
				l_env.Values[idx_name] = temp
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
		size, ok := s.(*IntVal).Get().(int)
		if !ok {
			return &RunError{cmd.Head.Oper.Line, cmd.Head.Oper.Start, "List size is not an integer"}
		}
		list := ListVal{make([]Val, size)}
		for i := range list.Stored {
			list.Stored[i] = &NullVal{}
		}
		env.Define(cmd.Def.Id, cmd.Def.Valtype)
		env.Put(cmd.Def.Id, cmd.Def.Valtype, &list, cmd.Head.Oper.Line, 0)
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
