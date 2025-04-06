package neu

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	Eof TokenType = iota
	Nl

	L_paren
	R_paren
	L_brace
	R_brace
	Comma

	Minus
	Plus
	Slash
	Star
	Sl_slash
	Percent

	Equal
	N_equal
	E_equal
	Greater
	G_equal
	Less
	L_equal

	Identifier
	Lit_str
	Lit_num
	Lit_char

	Type_num
	Type_str
	Type_bool
	Type_char
	Type_list
	Type_file

	Not
	And
	Or

	St_main
	Ed_main
	St_if
	Ed_if
	St_elif
	St_else
	St_while
	Ed_while
	St_for
	Ed_for

	C_print
	C_read

	Lit_true
	Lit_false

	Nul
)

func (t TokenType) String() string {
	switch t {
	case Eof:
		return "<EOF>"
	case Nl:
		return "<NL>"

	case L_paren:
		return "<L-PAR>"
	case R_paren:
		return "<R-PAR>"
	case L_brace:
		return "<L-BRA>"
	case R_brace:
		return "<R-BRA>"
	case Comma:
		return "<CM>"

	case Minus:
		return "<MIN>"
	case Plus:
		return "<PLS>"
	case Slash:
		return "<SLS>"
	case Star:
		return "<STA>"
	case Sl_slash:
		return "<SLSL>"
	case Percent:
		return "<PRC>"

	case Equal:
		return "<EQ>"
	case N_equal:
		return "<NEQ>"
	case E_equal:
		return "<EEQ>"
	case Greater:
		return "<GR>"
	case G_equal:
		return "<GEQ>"
	case Less:
		return "<LS>"
	case L_equal:
		return "<LEQ>"

	case Identifier:
		return "<ID>"
	case Lit_str:
		return "<LIT-STR>"
	case Lit_num:
		return "<LIT-NUM>"
	case Lit_char:
		return "<LIT-CHR>"

	case Type_num:
		return "<TYPE-NUM>"
	case Type_str:
		return "<TYPE-STR>"
	case Type_bool:
		return "<TYPE-LOG>"
	case Type_char:
		return "<TYPE-CHR>"
	case Type_list:
		return "<TYPE-LST>"
	case Type_file:
		return "<TYPE-FIL>"

	case Not:
		return "<NOT>"
	case And:
		return "<AND>"
	case Or:
		return "<OR>"

	case St_main:
		return "<PROG-ST>"
	case Ed_main:
		return "<PROG-ED>"
	case St_if:
		return "<IF-ST>"
	case Ed_if:
		return "<IF-ED>"
	case St_elif:
		return "<ELIF>"
	case St_else:
		return "<ELSE>"
	case St_while:
		return "<WHILE-ST>"
	case Ed_while:
		return "<WHILE-ED>"
	case St_for:
		return "<FOR-ST>"
	case Ed_for:
		return "<FOR-ED>"

	case C_print:
		return "<PRINT>"
	case C_read:
		return "<READ>"

	case Lit_true:
		return "<LIT-TRUE>"
	case Lit_false:
		return "<LIT-FALSE>"

	case Nul:
		return "<NUL>"
	default:
		return fmt.Sprintf("%d", t)
	}
}

type Token struct {
	Token_type TokenType
	Lexeme     string
	Literal    any
	Line       int
	Start      int
}

var keywords = map[string]TokenType{
	"PROG:":    St_main,
	":PROG":    Ed_main,
	"HA:":      St_if,
	":HA":      Ed_if,
	":DE-HA:":  St_elif,
	":NEM-HA:": St_else,
	"CIKLUS:":  St_while,
	":CIKLUS":  Ed_while,
	"ISM:":     St_for,
	":ISM":     Ed_for,
	"NOT":      Not,
	"AND":      And,
	"OR":       Or,
	"NUM:":     Type_num,
	"TXT:":     Type_str,
	"LOG:":     Type_bool,
	"KAR:":     Type_char,
	"FILE:":    Type_file,
	"KI:":      C_print,
	"BE:":      C_read,
	"IGAZ":     Lit_true,
	"HAMIS":    Lit_false,
	"SEMMI":    Nul,
}

type LexError struct {
	line int
	pos  int
	msg  string
}

func (e *LexError) Error() string {
	return fmt.Sprintf("[ERROR] Scanning Line %d Column %d: %s", e.line, e.pos, e.msg)
}

func next_match(line, check string, end *int) bool {
	endval := *end
	if endval+len(check) > len(line) {
		return false
	}
	matched := check == line[endval:endval+len(check)]
	if matched {
		*end += len(check) - 1
	}
	return matched
}

func match_until(line, limit string, end *int) (string, any) {
	endval := *end
	if endval+1 > len(line) {
		*end = len(line)
		return "", true
	}
	rest := line[endval+1:]
	idx := strings.Index(rest, limit)
	if idx < 0 {
		*end = len(line)
		return "", true
	}
	*end = endval + idx + 1
	return line[endval+1 : endval+idx+1], nil
}

func match_number(line string, end *int) (string, bool, int) {
	runes := []rune(line)
	var digits []rune
	endval := *end
	for ; endval < len(runes); endval++ {
		if !unicode.IsDigit(runes[endval]) {
			break
		}
		digits = append(digits, runes[endval])
	}
	if endval == len(runes) {
		*end = len(runes)
		return string(digits), true, -1
	}
	endindgs := []rune{' ', '+', '-', '*', '/', '%', ')', '<', '>', '=', '!', ','}
	if slices.Contains(endindgs, runes[endval]) {
		*end = endval - 1
		return string(digits), true, -1
	} else if runes[endval] == '.' {
		digits = append(digits, '.')
		endval++
		for ; endval < len(runes); endval++ {
			if !unicode.IsDigit(runes[endval]) {
				break
			}
			digits = append(digits, runes[endval])
		}
		*end = endval - 1
		return string(digits), false, -1
	} else {
		return string(digits), true, endval
	}
}

func is_name_char(char rune) bool {
	others := []rune{'-', '_', '?'}
	return unicode.IsLetter(char) || unicode.IsDigit(char) || slices.Contains(others, char)
}

func is_id_char(char rune) bool {
	others := []rune{'_', '?'}
	return unicode.IsLetter(char) || unicode.IsDigit(char) || slices.Contains(others, char)
}

func match_identifier(line string, end *int) (string, TokenType) {
	runes := []rune(line)
	var acc []rune
	endval := *end
	if runes[endval] == ':' {
		acc = append(acc, ':')
		endval++
	}
	for ; endval < len(runes) && is_name_char(runes[endval]); endval++ {
		acc = append(acc, runes[endval])
	}
	if endval < len(runes) {
		if runes[endval] == ':' {
			acc = append(acc, ':')
			endval++
		}
	}
	token_type, keyword := keywords[string(acc)]
	if !keyword {
		endval = *end
		acc = []rune{}
		for ; endval < len(runes) && is_id_char(runes[endval]); endval++ {
			acc = append(acc, runes[endval])
		}
		token_type = Identifier
	}
	*end = endval - 1
	return string(acc), token_type
}

func Lexer(line string, line_num int) ([]Token, []error) {
	var tokens []Token
	var err []error
	for current := 0; current < len(line); current++ {
		char := []rune(line)[current]
		var c_token Token
		if char == ' ' {
			continue
		}
		switch char {
		//single character
		case '(':
			c_token = Token{L_paren, "(", nil, line_num, current}
		case ')':
			c_token = Token{R_paren, ")", nil, line_num, current}
		case '[':
			c_token = Token{L_brace, "[", nil, line_num, current}
		case ']':
			c_token = Token{R_brace, "]", nil, line_num, current}
		case ',':
			c_token = Token{Comma, ",", nil, line_num, current}
		case '-':
			c_token = Token{Minus, "-", nil, line_num, current}
		case '+':
			c_token = Token{Plus, "+", nil, line_num, current}
		case '*':
			c_token = Token{Star, "*", nil, line_num, current}
		case '%':
			c_token = Token{Percent, "%", nil, line_num, current}
		//two character
		case '/':
			matched := next_match(line, "//", &current)
			if matched {
				c_token = Token{Sl_slash, "//", nil, line_num, current}
			} else {
				c_token = Token{Slash, "/", nil, line_num, current}
			}
		case '=':
			matched := next_match(line, "==", &current)
			if matched {
				c_token = Token{E_equal, "==", nil, line_num, current}
			} else {
				c_token = Token{Equal, "=", nil, line_num, current}
			}
		case '<':
			matched := next_match(line, "<=", &current)
			if matched {
				c_token = Token{L_equal, "<=", nil, line_num, current}
			} else {
				c_token = Token{Less, "<", nil, line_num, current}
			}
		case '>':
			matched := next_match(line, ">=", &current)
			if matched {
				c_token = Token{G_equal, ">=", nil, line_num, current}
			} else {
				c_token = Token{Greater, ">", nil, line_num, current}
			}
		case '!':
			matched := next_match(line, "!=", &current)
			if matched {
				c_token = Token{N_equal, "!=", nil, line_num, current}
			} else {
				err = append(err, &LexError{line_num, current, "Unexpected '!'. Did you mean !=? (remove the space)"})
			}
		case '#':
			//comment
			// ignore the rest of the characters
			current = len(line)
		//string/char literals
		case '"':
			value, e := match_until(line, "\"", &current)
			if e != nil {
				err = append(err, &LexError{line_num, current, "Unterminated string. Missing a: \""})
				continue
			}
			// value = value[1 : len(value)-1]
			c_token = Token{Lit_str, value, value, line_num, current}
		case '\'':
			value, e := match_until(line, "'", &current)
			if e != nil {
				err = append(err, &LexError{line_num, current, "Unterminated character. Missing a: '"})
				continue
			}
			if len(value) > 1 {
				err = append(err, &LexError{line_num, current, "Not a character: " + value})
				continue
			}
			// value = value[1 : len(value)-1]
			c_token = Token{Lit_char, value, []rune(value)[0], line_num, current}
		default:
			if unicode.IsSpace(char) {
				continue
			}
			if unicode.IsDigit(char) {
				value, integer, prob := match_number(line, &current)
				if prob != -1 {
					err = append(err, &LexError{line_num, current, "Not a valid number: " + line[current:prob+1]})
					current = prob
					continue
				}
				if integer {
					num, e := strconv.Atoi(value)
					if e != nil {
						err = append(err, &LexError{line_num, current, "Not a valid number. But it's my fault... " + value})
						continue
					}
					c_token = Token{Lit_num, value, num, line_num, current}
				} else {
					num, e := strconv.ParseFloat(value, 64)
					if e != nil {
						err = append(err, &LexError{line_num, current, "Not a valid number. But it's my fault... " + value})
						continue
					}
					c_token = Token{Lit_num, value, num, line_num, current}
				}
			} else if unicode.IsLetter(char) || char == ':' {
				value, token_type := match_identifier(line, &current)
				if value == "" {
					fmt.Println(line_num)
					current++
					continue
				}
				if token_type == Lit_true || token_type == Lit_false {
					c_token = Token{token_type, value, token_type == Lit_true, line_num, current}
				} else {
					c_token = Token{token_type, value, nil, line_num, current}
				}
			} else {
				err = append(err, &LexError{line_num, current, "Unrecognized character"})
			}
		}
		if c_token.Line != 0 {
			tokens = append(tokens, c_token)
		}
	}
	if len(tokens) > 0 {
		tokens = append(tokens, Token{Nl, "NewLine", nil, line_num, tokens[len(tokens)-1].Start + 1})
	}

	return tokens, err
}
