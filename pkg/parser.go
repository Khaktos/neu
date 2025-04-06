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
	left  *TreeNode
	oper  Token
	right *TreeNode
}

type CommandType int

const (
	prog_cmd CommandType = iota
	print_cmd
	read_cmd
	vardef_cmd
	assign_cmd
	if_cmd
	for_cmd
	while_cmd
	block
	expr_cmd
)

func (c CommandType) String() string {
	switch c {
	case prog_cmd:
		return "<PROG>"
	case print_cmd:
		return "<PRINT>"
	case read_cmd:
		return "<READ>"
	case vardef_cmd:
		return "<V-DEF>"
	case assign_cmd:
		return "<ASSIGN>"
	case if_cmd:
		return "<IF>"
	case for_cmd:
		return "<FOR>"
	case while_cmd:
		return "<WHILE>"
	case block:
		return "<BLOCK>"
	case expr_cmd:
		return "<EXPR>"
	default:
		return fmt.Sprintf("%d", c)
	}
}

type Command struct {
	stmt_type CommandType
	head      *TreeNode
	body      []*Command
	id        string
	vtype     TokenType
}

type ParseError struct {
	line int
	pos  int
	msg  string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("[ERROR] Parsing Line %d Column %d: %s", e.line, e.pos+1, e.msg)
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
		// if !(*current+1 >= len(tokens)) {
		*current++
		// }
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

func parse_primary(tokens []Token, current *int) (*TreeNode, error) {
	if match_token(tokens, current, Lit_false, Lit_true, Lit_char, Lit_num, Lit_str, Nul, Identifier) {
		literal := tokens[*current-1]
		return &TreeNode{oper: literal}, nil
	}
	if match_token(tokens, current, L_paren) {
		paren := tokens[*current-1]
		expr, err := parse_expression(tokens, current)
		if err != nil {
			return nil, err
		}
		if match_token(tokens, current, R_paren) {
			return &TreeNode{left: expr, oper: paren}, nil
		} else {
			return nil, &ParseError{paren.Line, paren.Start, "Unclosed parenthesis"}
		}
		// end := find_last(tokens, R_paren)
		// fmt.Println(*current, end)
		// if end == -1 {
		// 	return nil, &ParseError{paren.Line, paren.Start, "Unclosed parenthesis"}
		// }
		// sub_curr := 0
		// expr, err := parse_expression(tokens[*current:end], &sub_curr)
		// if err != nil {
		// 	return nil, err
		// }
		// *current = end
		// return &TreeNode{left: expr, oper: paren}, nil
	}
	return nil, &ParseError{tokens[*current].Line, tokens[*current].Start, "Primary token missing"}
}

func parse_unary(tokens []Token, current *int) (*TreeNode, error) {
	if match_token(tokens, current, Not, Minus) {
		operator := tokens[*current-1]
		right, err := parse_unary(tokens, current)
		return &TreeNode{oper: operator, right: right}, err
	}
	return parse_primary(tokens, current)
}

func parse_factor(tokens []Token, current *int) (*TreeNode, error) {
	expr, err := parse_unary(tokens, current)
	// if err != nil {
	// return nil, err
	// }
	for match_token(tokens, current, Star, Slash, Sl_slash, Percent) {
		operator := tokens[*current-1]
		right, e := parse_unary(tokens, current)
		expr = &TreeNode{left: expr, right: right, oper: operator}
		if e != nil {
			err = e
			// return expr, e
		}
	}
	return expr, err
}

func parse_term(tokens []Token, current *int) (*TreeNode, error) {
	expr, err := parse_factor(tokens, current)
	// if err != nil {
	// return nil, err
	// }
	for match_token(tokens, current, Plus, Minus) {
		operator := tokens[*current-1]
		right, e := parse_factor(tokens, current)
		expr = &TreeNode{left: expr, right: right, oper: operator}
		if e != nil {
			err = e
			// return expr, e
		}
	}
	return expr, err
}

func parse_comparison(tokens []Token, current *int) (*TreeNode, error) {
	expr, err := parse_term(tokens, current)
	// if err != nil {
	// return nil, err
	// }
	for match_token(tokens, current, Greater, G_equal, Less, L_equal) {
		operator := tokens[*current-1]
		right, e := parse_term(tokens, current)
		expr = &TreeNode{left: expr, right: right, oper: operator}
		if e != nil {
			err = e
			// return expr, e
		}
	}
	return expr, err
}

func parse_equality(tokens []Token, current *int) (*TreeNode, error) {
	expr, err := parse_comparison(tokens, current)
	// if err != nil {
	// return nil, err
	// }
	for match_token(tokens, current, E_equal, N_equal) {
		operator := tokens[*current-1]
		right, e := parse_comparison(tokens, current)
		expr = &TreeNode{left: expr, right: right, oper: operator}
		if e != nil {
			err = e
			// return expr, e
		}
	}
	return expr, err
}

func parse_expression(tokens []Token, current *int) (*TreeNode, error) {
	if len(tokens) == 0 {
		return nil, &ParseError{0, 0, "Primary token missing"}
	}
	expr, err := parse_equality(tokens, current)
	// if err != nil {
	// return nil, err
	// }
	for match_token(tokens, current, And, Or) {
		operator := tokens[*current-1]
		right, e := parse_equality(tokens, current)
		expr = &TreeNode{left: expr, right: right, oper: operator}
		if e != nil {
			err = e
			// return expr, e
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
	return &Command{block, nil, body, "", Nul}, nil
}

func parse_if(tokens []Token, current *int) (*Command, error) {
	// find Ed_if
	end_if := *current
	for ; end_if < len(tokens) && !(tokens[end_if].Token_type == Ed_if); end_if++ {
	}
	if end_if == len(tokens) {
		return nil, &ParseError{tokens[end_if-1].Line, tokens[end_if-1].Start, "HA tag not closed"}
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

	return &Command{if_cmd, head, body, "", Nul}, nil
}

func parse_loop(tokens []Token, current *int, loop_type CommandType) (*Command, error) {
	if loop_type != for_cmd && loop_type != while_cmd {
		panic("Unknown loop type")
	}
	msg_type := "ISM"
	if loop_type == while_cmd {
		msg_type = "CIKLUS"
	}
	message := fmt.Sprintf("%s tag not closed", msg_type)
	end_for := *current
	for ; end_for < len(tokens) && !(tokens[end_for].Token_type == Ed_for); end_for++ {
	}
	if end_for == len(tokens) {
		return nil, &ParseError{tokens[end_for-1].Line, tokens[end_for-1].Start, message}
	}
	head, err := parse_expression(tokens, current)
	var idx_def *Command
	if match_token(tokens, current, Comma) {
		if match_token(tokens, current, Identifier) {
			// index variable
			// define new num variable
			// set value to zero
			idx_par := Token{Lit_num, "<IDX-PAR>", 0, tokens[*current].Line, tokens[*current].Start + 1}
			idx_literal := &TreeNode{oper: idx_par}
			idx_def = &Command{vardef_cmd, idx_literal, nil, tokens[*current-1].Lexeme, Type_num}
		} else {
			return nil, &ParseError{tokens[*current].Line, tokens[*current].Start, "Expected indexing variable"}
		}
	}
	err = expect_nl(tokens, current)
	if err != nil {
		return nil, err
	}
	for_block, err := parse_block(tokens, current, Ed_for)
	if err != nil {
		return nil, err
	}
	match_token(tokens, current, Ed_for)
	err = expect_nl(tokens, current)
	if err != nil {
		return nil, err
	}
	var body []*Command
	body = append(body, idx_def)
	body = append(body, for_block)

	return &Command{loop_type, head, body, "", Nul}, nil
}

func parse_print(tokens []Token, current *int) (*Command, error) {
	var body []*Command
	expr, err := parse_expression(tokens, current)
	if err != nil {
		return nil, err
	}

	body = append(body, &Command{print_cmd, expr, nil, "", Nul})

	for match_token(tokens, current, Comma) {
		expr, err = parse_expression(tokens, current)
		if err != nil {
			return nil, err
		}
		space := TreeNode{oper: Token{Lit_str, "<PRINT-SEP>", " ", tokens[*current].Line, tokens[*current].Start}}
		body = append(body, &Command{print_cmd, &space, nil, "", Nul})
		body = append(body, &Command{print_cmd, expr, nil, "", Nul})
	}
	err = expect_nl(tokens, current)
	if err != nil {
		return nil, err
	}

	newline := TreeNode{oper: Token{Lit_str, "<PRINT-NL>", "\n", tokens[*current-1].Line, tokens[*current-1].Start}}
	body = append(body, &Command{print_cmd, &newline, nil, "", Nul})
	return &Command{block, nil, body, "", Nul}, nil
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
		return &Command{read_cmd, head, nil, "", Nul}, nil
	}
	//declaration
	if match_token(tokens, current, Type_bool, Type_char, Type_num, Type_str) {
		vtype := tokens[*current-1].Token_type
		if match_token(tokens, current, Identifier) {
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
			return &Command{vardef_cmd, head, nil, id, vtype}, nil

		} else {
			err := &ParseError{tokens[*current].Line, tokens[*current].Start, "Identifier expected"}
			sync_to_next_cmd(tokens, current)
			return nil, err
		}
	}
	//assignment
	if match_token_seq(tokens, current, []TokenType{Identifier, Equal}) {
		id := tokens[*current-2].Lexeme
		head, err := parse_expression(tokens, current)
		if err != nil {
			return nil, err
		}
		err = expect_nl(tokens, current)
		if err != nil {
			return nil, err
		}
		return &Command{assign_cmd, head, nil, id, Nul}, nil
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
		cmd, err := parse_loop(tokens, current, for_cmd)
		if err != nil {
			return nil, err
		}
		return cmd, nil
	}
	//while
	if match_token(tokens, current, St_while) {
		cmd, err := parse_loop(tokens, current, while_cmd)
		if err != nil {
			return nil, err
		}
		return cmd, nil
	}

	//fallback -> naked expression
	{
		head, err := parse_expression(tokens, current)
		if err != nil {
			return nil, err
		}
		err = expect_nl(tokens, current)
		if err != nil {
			return nil, err
		}
		return &Command{expr_cmd, head, nil, "", Nul}, nil
	}
}

func sync_to_next_cmd(tokens []Token, current *int) {
	for ; !match_token(tokens, current, Nl); *current++ {
		// fmt.Println("syncing", tokens[*current])
		if *current >= len(tokens) {
			break
		}
	}
	*current--
}

func Parser(tokens []Token) (*Command, []error) {
	var err []error
	//find main PROG tags
	start, end := -1, -1
	for i := 0; tokens[i].Token_type != Eof; i++ {
		if tokens[i].Token_type == St_main {
			start = i + 1
		}
		if tokens[i].Token_type == Ed_main {
			end = i
		}
	}
	if start < 0 {
		err = append(err, &ParseError{0, 0, "No PROG tag found"})
		start = 0
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
		head = &TreeNode{oper: main_tokens[current-1]}
	}

	if head != nil && !match_token(main_tokens, &current, Nl) {
		err = append(err, &ParseError{main_tokens[current].Line, main_tokens[current].Start, "New line expected"})
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
	return &Command{prog_cmd, head, body, "", Nul}, err
}
