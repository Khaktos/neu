package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"khaktos.digital/neu/pkg"
)

func main() {
	had_error := false
	debug_mode := false
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Please provide a .neu file as input")
		os.Exit(64)
	}
	var file *os.File
	var err error
	if len(args) == 1 {
		file, err = os.Open(args[0])
	}
	if len(args) == 2 {
		if args[0] == "-d" {
			debug_mode = true
		}
		file, err = os.Open(args[1])
	}
	if len(args) > 2 {
		fmt.Println("Multiple arguments passed. This is currently unsupported.")
	}
	if err != nil {
		panic(err)
	}

	//reading
	scanner := bufio.NewScanner(file)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	//lexing
	var tokens []neu.Token
	var errs []error
	for ind, line := range lines {

		if strings.TrimSpace(line) == "" {
			continue
		}
		line_tokens, err := neu.Lexer(line, ind+1)
		if len(err) != 0 {
			errs = append(errs, err...)
			had_error = true
		}
		tokens = append(tokens, line_tokens...)
	}
	tokens = tokens[:len(tokens)-1]
	tokens = append(tokens, neu.Token{Token_type: neu.Eof, Lexeme: "EOF", Literal: nil, Line: len(lines) + 1})

	if had_error {
		for _, v := range errs {
			fmt.Fprintf(os.Stderr, "%s\n", v)
		}
		os.Exit(1)
	}
	if debug_mode {
		fmt.Println("Debug info:\n-------------")
		for _, tok := range tokens {
			fmt.Printf("%+v\n", tok)
		}
	}

	//parsing
	program, perr := neu.Parser(tokens)
	if len(perr) != 0 {
		for _, v := range perr {
			fmt.Fprintf(os.Stderr, "%s\n", v)
		}
		os.Exit(1)
	}
	if debug_mode {
		program.Print()
		fmt.Println()
		fmt.Println("------------")
	}

	//running
	program.Interpret(nil)
}
