package neu

import (
	"fmt"
	"slices"
)

// TreeNode will be used for all AST nodes:
// Binary: both children set
// Unary: only right child set
// Group: only left child set
// Literal: no child set
type TreeNode struct {
	Left  *TreeNode
	Oper  Token
	Right *TreeNode
}

type CommandType int

const (
	Prog_cmd CommandType = iota
	Print_cmd
	Read_cmd
	Vardef_cmd
	Listdef_cmd
	Assign_cmd
	Index_cmd
	If_cmd
	For_cmd
	While_cmd
	Block
	Expr_cmd
)

func (c CommandType) String() string {
	switch c {
	case Prog_cmd:
		return "<PROG>"
	case Print_cmd:
		return "<PRINT>"
	case Read_cmd:
		return "<READ>"
	case Vardef_cmd:
		return "<VAR-DEF>"
	case Listdef_cmd:
		return "<LIST-DEF>"
	case Assign_cmd:
		return "<ASSIGN>"
	case Index_cmd:
		return "<INDEX>"
	case If_cmd:
		return "<IF>"
	case For_cmd:
		return "<FOR>"
	case While_cmd:
		return "<WHILE>"
	case Block:
		return "<BLOCK>"
	case Expr_cmd:
		return "<EXPR>"
	default:
		return fmt.Sprintf("%d", c)
	}
}

type Command struct {
	Stmt_type CommandType
	Head      *TreeNode
	Body      []*Command
	Def       Variable
}

type ParseError struct {
	line int
	pos  int
	msg  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("[ERROR] Parsing: Line %d Column %d: %s", e.line, e.pos+1, e.msg)
}

func expect_nl(tokens []Token, current *int) error {
	if !(match_token(tokens, current, Nl)) {
		re := &ParseError{tokens[*current].Line, tokens[*current].Start, "New line expected"}
		sync_to_next_cmd(tokens, current)
		return re
	}
	return nil
}

func match_token(tokens []Token, current *int, t_match ...TokenType) bool {
	if *current >= len(tokens) {
		return false
	}
	if slices.Contains(t_match, tokens[*current].Token_type) {
		*current++
		return true
	} else {
		return false
	}
}

func match_token_seq(tokens []Token, current *int, seq []TokenType) bool {
	reset := *current
	succ := true
	for _, typ := range seq {
		if !match_token(tokens, current, typ) {
			succ = false
			*current = reset
			break
		}
	}
	return succ
}

func contains_type(tokens []Token, typ TokenType) bool {
	contain := false
	for _, token := range tokens {
		if token.Token_type == typ {
			contain = true
			break
		}
	}
	return contain
}

func find_last(tokens []Token, typ TokenType) int {
	for i, v := range slices.Backward(tokens) {
		if v.Token_type == typ {
			return i
		}
	}
	return -1
}

// TODO: add support for multi dimension arrays
func parse_indexing(tokens []Token, current *int) (*TreeNode, error) {
	ident := tokens[*current-1]
	ret := &TreeNode{Oper: ident}
	if match_token(tokens, current, L_brace) {
		brace := tokens[*current-1]
		expr, err := parse_expression(tokens, current)
		if err != nil {
			return nil, err
		}
		if match_token(tokens, current, R_brace) {
			return &TreeNode{Left: ret, Oper: brace, Right: expr}, nil
		} else {
			return nil, &ParseError{brace.Line, brace.Start, "Unclosed indexing"}
		}
	}
	return ret, nil
}

func parse_primary(tokens []Token, current *int) (*TreeNode, error) {
	if match_token(tokens, current, Identifier) {
		indexed, err := parse_indexing(tokens, current)
		if err != nil {
			return nil, err
		}
		return indexed, nil
	}
	if match_token(tokens, current, Lit_false, Lit_true, Lit_char, Lit_num, Lit_str, Nul) {
		literal := tokens[*current-1]
		return &TreeNode{Oper: literal}, nil
	}
	if match_token(tokens, current, L_paren) {
		paren := tokens[*current-1]
		expr, err := parse_expression(tokens, current)
		if err != nil {
			return nil, err
		}
		if match_token(tokens, current, R_paren) {
			return &TreeNode{Left: expr, Oper: paren}, nil
		} else {
			return nil, &ParseError{paren.Line, paren.Start, "Unclosed parenthesis"}
		}
	}
	return nil, &ParseError{tokens[*current].Line, tokens[*current].Start, "Primary token missing"}
}

func parse_unary(tokens []Token, current *int) (*TreeNode, error) {
	if match_token(tokens, current, Not, Minus) {
		operator := tokens[*current-1]
		right, err := parse_unary(tokens, current)
		return &TreeNode{Oper: operator, Right: right}, err
	}
	return parse_primary(tokens, current)
}

func parse_factor(tokens []Token, current *int) (*TreeNode, error) {
	expr, err := parse_unary(tokens, current)
	for match_token(tokens, current, Star, Slash, Sl_slash, Percent) {
		operator := tokens[*current-1]
		right, e := parse_unary(tokens, current)
		expr = &TreeNode{Left: expr, Right: right, Oper: operator}
		if e != nil {
			err = e
		}
	}
	return expr, err
}

func parse_term(tokens []Token, current *int) (*TreeNode, error) {
	expr, err := parse_factor(tokens, current)
	for match_token(tokens, current, Plus, Minus) {
		operator := tokens[*current-1]
		right, e := parse_factor(tokens, current)
		expr = &TreeNode{Left: expr, Right: right, Oper: operator}
		if e != nil {
			err = e
		}
	}
	return expr, err
}

func parse_comparison(tokens []Token, current *int) (*TreeNode, error) {
	expr, err := parse_term(tokens, current)
	for match_token(tokens, current, Greater, G_equal, Less, L_equal) {
		operator := tokens[*current-1]
		right, e := parse_term(tokens, current)
		expr = &TreeNode{Left: expr, Right: right, Oper: operator}
		if e != nil {
			err = e
		}
	}
	return expr, err
}

func parse_equality(tokens []Token, current *int) (*TreeNode, error) {
	expr, err := parse_comparison(tokens, current)
	for match_token(tokens, current, E_equal, N_equal) {
		operator := tokens[*current-1]
		right, e := parse_comparison(tokens, current)
		expr = &TreeNode{Left: expr, Right: right, Oper: operator}
		if e != nil {
			err = e
		}
	}
	return expr, err
}

func parse_expression(tokens []Token, current *int) (*TreeNode, error) {
	if len(tokens) == 0 {
		return nil, &ParseError{0, 0, "Primary token missing"}
	}
	expr, err := parse_equality(tokens, current)
	for match_token(tokens, current, And, Or) {
		operator := tokens[*current-1]
		right, e := parse_equality(tokens, current)
		expr = &TreeNode{Left: expr, Right: right, Oper: operator}
		if e != nil {
			err = e
		}
	}
	return expr, err
}

func parse_block(tokens []Token, current *int, closers ...TokenType) (*Command, error) {
	var body []*Command
	for *current < len(tokens) &&
		!(slices.Contains(closers, tokens[*current].Token_type)) {
		cmd, err := parse_command(tokens, current)
		if err != nil {
			return nil, err
		}
		body = append(body, cmd)
	}
	return &Command{Block, nil, body, NoVar}, nil
}

func parse_if(tokens []Token, current *int) (*Command, error) {
	// find Ed_if
	end_if := slices.IndexFunc(tokens[*current:], func(e Token) bool { return e.Token_type == Ed_if })
	if end_if < 0 {
		return nil, &ParseError{tokens[*current].Line, tokens[*current].Start, "HA tag not closed"}
	}

	head, err := parse_expression(tokens, current)
	if err != nil {
		return nil, err
	}
	err = expect_nl(tokens, current)
	if err != nil {
		return nil, err
	}
	true_cmd, err := parse_block(tokens, current, Ed_if, St_elif, St_else)
	if err != nil {
		return nil, err
	}

	//TODO: possible easier implementation?
	var false_cmd *Command
	if tokens[*current].Token_type != Ed_if {
		//check if elif: -> parse as if
		//      if else: -> just parse body
		if match_token(tokens, current, St_elif) {
			cmd, err := parse_if(tokens, current)
			if err != nil {
				return nil, err
			}
			false_cmd = cmd
		} else if match_token(tokens, current, St_else) {
			err = expect_nl(tokens, current)
			if err != nil {
				return nil, err
			}
			cmd, err := parse_block(tokens, current, Ed_if)
			if err != nil {
				return nil, err
			}
			false_cmd = cmd
			match_token(tokens, current, Ed_if)
			err = expect_nl(tokens, current)
			if err != nil {
				return nil, err
			}

		} else {
			panic("Something went wrong while parsing IF")
		}
	} else {
		match_token(tokens, current, Ed_if)
		err = expect_nl(tokens, current)
		if err != nil {
			return nil, err
		}
	}
	body := []*Command{true_cmd, false_cmd}

	return &Command{If_cmd, head, body, NoVar}, nil
}

func parse_loop(tokens []Token, current *int, loop_type CommandType) (*Command, error) {
	if loop_type != For_cmd && loop_type != While_cmd {
		panic("Unknown loop type")
	}
	msg_type := "ISM"
	end_type := Ed_for
	if loop_type == While_cmd {
		msg_type = "CIKLUS"
		end_type = Ed_while
	}
	message := fmt.Sprintf("%s tag not closed", msg_type)
	end_loop := slices.IndexFunc(tokens[*current:], func(e Token) bool { return e.Token_type == end_type })
	if end_loop < 0 {
		return nil, &ParseError{tokens[*current].Line, tokens[*current].Start, message}
	}
	head, err := parse_expression(tokens, current)
	if err != nil {
		return nil, err
	}
	var idx_def *Command
	if match_token(tokens, current, Comma) {
		if match_token(tokens, current, Identifier) {
			// index variable
			// define new num variable
			// set value to zero
			idx_par := Token{Lit_num, "<IDX-PAR>", &IntVal{0}, tokens[*current].Line, tokens[*current].Start + 1}
			idx_literal := &TreeNode{Oper: idx_par}
			idx_def = &Command{Vardef_cmd, idx_literal, nil, Variable{tokens[*current-1].Lexeme, BaseNum}}
		} else {
			return nil, &ParseError{tokens[*current].Line, tokens[*current].Start, "Expected indexing variable"}
		}
	}
	err = expect_nl(tokens, current)
	if err != nil {
		return nil, err
	}
	loop_block, err := parse_block(tokens, current, end_type)
	if err != nil {
		return nil, err
	}
	match_token(tokens, current, end_type)
	err = expect_nl(tokens, current)
	if err != nil {
		return nil, err
	}
	var body []*Command
	body = append(body, idx_def)
	body = append(body, loop_block)

	return &Command{loop_type, head, body, NoVar}, nil
}

func parse_print(tokens []Token, current *int) (*Command, error) {
	var body []*Command
	expr, err := parse_expression(tokens, current)
	if err != nil {
		return nil, err
	}

	body = append(body, &Command{Print_cmd, expr, nil, NoVar})

	for match_token(tokens, current, Comma) {
		expr, err = parse_expression(tokens, current)
		if err != nil {
			return nil, err
		}
		space := TreeNode{Oper: Token{Lit_str, "<PRINT-SEP>", &StrVal{" "}, tokens[*current].Line, tokens[*current].Start}}
		body = append(body, &Command{Print_cmd, &space, nil, NoVar})
		body = append(body, &Command{Print_cmd, expr, nil, NoVar})
	}
	err = expect_nl(tokens, current)
	if err != nil {
		return nil, err
	}

	newline := TreeNode{Oper: Token{Lit_str, "<PRINT-NL>", &StrVal{"\n"}, tokens[*current-1].Line, tokens[*current-1].Start}}
	body = append(body, &Command{Print_cmd, &newline, nil, NoVar})
	return &Command{Block, nil, body, NoVar}, nil
}

func parse_command(tokens []Token, current *int) (*Command, error) {
	//printing
	if match_token(tokens, current, C_print) {
		return parse_print(tokens, current)
	}
	//reading
	if match_token(tokens, current, C_read) {
		head, err := parse_primary(tokens, current)
		if err != nil {
			return nil, err
		}
		err = expect_nl(tokens, current)
		if err != nil {
			return nil, err
		}
		return &Command{Read_cmd, head, nil, NoVar}, nil
	}
	//declaration
	if match_token(tokens, current, Type_bool, Type_char, Type_num, Type_str) {
		dec_type_tok := tokens[*current-1].Token_type
		vtype := BaseNum
		switch dec_type_tok {
		case Type_bool:
			vtype = BaseBool
		case Type_char:
			vtype = BaseChar
		case Type_str:
			vtype = BaseStr
		}
		if match_token_seq(tokens, current, []TokenType{Colon, Identifier}) {
			id := tokens[*current-1].Lexeme
			//check if setter expression is present -> add it to head
			var head *TreeNode
			var err error
			if match_token(tokens, current, Equal) {
				head, err = parse_expression(tokens, current)
			}
			if err != nil {
				return nil, err
			}
			err = expect_nl(tokens, current)
			if err != nil {
				return nil, err
			}
			return &Command{Vardef_cmd, head, nil, Variable{id, vtype}}, nil

		} else {
			err := &ParseError{tokens[*current].Line, tokens[*current].Start, "Identifier expected"}
			sync_to_next_cmd(tokens, current)
			return nil, err
		}
	}
	//list declaration
	if match_token(tokens, current, Type_list) {
		curr_type, err := parse_list_type(tokens, current)
		if err != nil {
			return nil, err
		}
		if !match_token_seq(tokens, current, []TokenType{Colon, Identifier, L_brace}) {
			err = &ParseError{tokens[*current].Line, tokens[*current].Start, "Syntax error in list declaration a"}
			sync_to_next_cmd(tokens, current)
			return nil, err
		}
		id := tokens[*current-2].Lexeme
		size, err := parse_expression(tokens, current)
		if err != nil {
			sync_to_next_cmd(tokens, current)
			return nil, err
		}
		if !match_token(tokens, current, R_brace) {
			err = &ParseError{tokens[*current].Line, tokens[*current].Start, "Syntax error in list declaration"}
			sync_to_next_cmd(tokens, current)
			return nil, err
		}
		err = expect_nl(tokens, current)
		if err != nil {
			return nil, err
		}
		return &Command{Listdef_cmd, size, nil, Variable{id, curr_type}}, nil
	}

	//assignment
	// if match_token_seq(tokens, current, []TokenType{Identifier, Equal}) {
	// 	id := tokens[*current-2].Lexeme
	// 	head, err := parse_expression(tokens, current)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	err = expect_nl(tokens, current)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return &Command{Assign_cmd, head, nil, id, Nul}, nil
	// }
	if match_token(tokens, current, Identifier) {
		indexed, err := parse_indexing(tokens, current)
		if err != nil {
			return nil, err
		}
		// fmt.Println(indexed)
		if !match_token(tokens, current, Equal) {
			panic("Unreachable?")
		}
		head, err := parse_expression(tokens, current)
		if err != nil {
			return nil, err
		}
		err = expect_nl(tokens, current)
		if err != nil {
			return nil, err
		}
		switch indexed.Oper.Token_type {
		case Identifier:
			return &Command{Assign_cmd, head, nil, Variable{indexed.Oper.Lexeme, ValueType{}}}, nil
		case L_brace:
			return &Command{Assign_cmd, head, []*Command{
				{Index_cmd, indexed.Right, nil, NoVar},
			}, Variable{indexed.Left.Oper.Lexeme, ValueType{}}}, nil
		}
		// &Command{List_cmd, head, nil, Variable{indexed.Oper.Left.Lexeme, ValueType{}}}, nil
		// this is jank AF
		// different assignment thing is needed in the interpretation part
	}
	//if
	if match_token(tokens, current, St_if) {
		cmd, err := parse_if(tokens, current)
		if err != nil {
			return nil, err
		}
		return cmd, nil
	}
	//for
	if match_token(tokens, current, St_for) {
		cmd, err := parse_loop(tokens, current, For_cmd)
		if err != nil {
			return nil, err
		}
		return cmd, nil
	}
	//while
	if match_token(tokens, current, St_while) {
		cmd, err := parse_loop(tokens, current, While_cmd)
		if err != nil {
			return nil, err
		}
		return cmd, nil
	}

	return nil, &ParseError{tokens[*current].Line, tokens[*current].Start, "Unrecognized command type: " + tokens[*current].Lexeme}

	// //fallback -> naked expression
	// {
	// 	head, err := parse_expression(tokens, current)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	err = expect_nl(tokens, current)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return &Command{Expr_cmd, head, nil, "", Nul}, nil
	// }
}

func parse_list_type(tokens []Token, current *int) (ValueType, error) {
	elem_type := ValueType{List_type, []ValueType{}}
	if !match_token(tokens, current, L_brace) {
		err := &ParseError{tokens[*current].Line, tokens[*current].Start, "'[' expected"}
		return ValueType{}, err
	}
	if match_token(tokens, current, Type_list) {
		el, err := parse_list_type(tokens, current)
		if err != nil {
			return ValueType{}, err
		}
		elem_type.SubType = append(elem_type.SubType, el)
	} else if match_token(tokens, current, Type_bool, Type_char, Type_num, Type_str) {
		elem_type.SubType = append(elem_type.SubType, token_to_type[tokens[*current-1].Token_type])
	} else {
		err := &ParseError{tokens[*current].Line, tokens[*current].Start, "Unsupported type for list elements"}
		return ValueType{}, err
	}
	if !match_token(tokens, current, R_brace) {
		err := &ParseError{tokens[*current].Line, tokens[*current].Start, "']' expected"}
		return ValueType{}, err
	}
	return elem_type, nil
}

func sync_to_next_cmd(tokens []Token, current *int) {
	for ; !match_token(tokens, current, Nl); *current++ {
		// fmt.Println("syncing", tokens[*current])
		if *current >= len(tokens) {
			break
		}
	}
	// *current--
}

func Pre_parse(tokens []Token) ([]Token, error) {
	var new_tokens []Token
	skip := false
	for idx, token := range tokens {
		if skip {
			skip = false
			continue
		}
		prev_colon := tokens[max(idx-1, 0)].Token_type == Colon
		next_colon := tokens[min(idx+1, len(tokens)-1)].Token_type == Colon
		curr := token
		err := &ParseError{curr.Line, curr.Start, "Tag not opened or closed"}
		switch token.Token_type {
		case Main:
			if next_colon {
				curr = Token{St_main, "PROG:", nil, curr.Line, curr.Start}
				skip = true
			} else if prev_colon {
				new_tokens = new_tokens[:len(new_tokens)-1]
				curr = Token{Ed_main, ":PROG", nil, curr.Line, curr.Start - 1}
			} else {
				return nil, err
			}
		case If:
			if next_colon {
				curr = Token{St_if, "HA:", nil, curr.Line, curr.Start}
				skip = true
			} else if prev_colon {
				new_tokens = new_tokens[:len(new_tokens)-1]
				curr = Token{Ed_if, ":HA", nil, curr.Line, curr.Start - 1}
			} else {
				return nil, err
			}
		case Elif:
			if next_colon && prev_colon {
				new_tokens = new_tokens[:len(new_tokens)-1]
				curr = Token{St_elif, ":DE-HA:", nil, curr.Line, curr.Start - 1}
				skip = true
			} else {
				return nil, err
			}
		case Else:
			if next_colon && prev_colon {
				new_tokens = new_tokens[:len(new_tokens)-1]
				curr = Token{St_else, ":NEM-HA:", nil, curr.Line, curr.Start - 1}
				skip = true
			} else {
				return nil, err
			}
		case While:
			if next_colon {
				curr = Token{St_while, "CIKLUS:", nil, curr.Line, curr.Start}
				skip = true
			} else if prev_colon {
				new_tokens = new_tokens[:len(new_tokens)-1]
				curr = Token{Ed_while, ":CIKLUS", nil, curr.Line, curr.Start - 1}
			} else {
				return nil, err
			}
		case For:
			if next_colon {
				curr = Token{St_for, "ISM:", nil, curr.Line, curr.Start}
				skip = true
			} else if prev_colon {
				new_tokens = new_tokens[:len(new_tokens)-1]
				curr = Token{Ed_for, ":ISM", nil, curr.Line, curr.Start - 1}
			} else {
				return nil, err
			}
		case Print:
			if next_colon {
				curr = Token{C_print, "KI:", nil, curr.Line, curr.Start}
				skip = true
			} else {
				return nil, err
			}
		case Read:
			if next_colon {
				curr = Token{C_read, "BE:", nil, curr.Line, curr.Start}
				skip = true
			} else {
				return nil, err
			}
		}
		new_tokens = append(new_tokens, curr)
	}
	return new_tokens, nil
}

func Parser(tokens []Token) (*Command, []error) {
	var err []error

	//find main PROG tags
	start := slices.IndexFunc(tokens, func(e Token) bool { return e.Token_type == St_main })
	end := slices.IndexFunc(tokens, func(e Token) bool { return e.Token_type == Ed_main })
	if start < 0 {
		err = append(err, &ParseError{0, 0, "No PROG tag found"})
		start = 0
	} else {
		start++
	}
	if end < 0 {
		etok := max(0, len(tokens)-2)
		err = append(err, &ParseError{tokens[etok].Line, tokens[etok].Start, "PROG tag not closed"})
		end = len(tokens)
	}

	main_tokens := tokens[start:end]
	current := 0
	var head *TreeNode

	if !match_token(main_tokens, &current, Identifier) {
		err = append(err, &ParseError{main_tokens[current].Line, main_tokens[current].Start, "No PROG indentifier"})
		// sync_to_next_cmd(main_tokens, &current)
		current++
	} else {
		head = &TreeNode{Oper: main_tokens[current-1]}
	}

	if head != nil && !match_token(main_tokens, &current, Nl) {
		err = append(err, &ParseError{main_tokens[current].Line, main_tokens[current].Start, "New line expected"})
		// sync_to_next_cmd(main_tokens, &current)
		current++
	}

	var body []*Command
	for current < len(main_tokens) && main_tokens[current].Token_type != Eof {
		cmd, e := parse_command(main_tokens, &current)
		if e != nil {
			err = append(err, e)
			sync_to_next_cmd(main_tokens, &current)
			continue
		}
		body = append(body, cmd)
	}
	return &Command{Prog_cmd, head, body, NoVar}, err
}
