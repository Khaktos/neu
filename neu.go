package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"khaktos.digital/neu/pkg"
)

func main() {

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of neu:\n\tneu <options> <neu program>\nOptions:\n")
		flag.PrintDefaults()
	}
	debug_flag := flag.Bool("d", false, "Debug mode")

	had_error := false
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		fmt.Println("Please provide a .neu file as input")
		os.Exit(64)
	}
	var file *os.File
	var err error

	if len(args) == 1 {
		file, err = os.Open(args[0])
	}
	if len(args) > 1 {
		fmt.Println("Multiple arguments passed. This is currently unsupported.")
		os.Exit(1)
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
	if *debug_flag {
		fmt.Println("-------------\nDebug info:")
		fmt.Println("\nLexer output:")
		for i, tok := range tokens {
			fmt.Printf("%d: %+v\n", i, tok)
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
	if *debug_flag {
		fmt.Println("\nParser output:")
		program.Print("")
		fmt.Println()
		fmt.Println("------------")
	}

	//running
	program.Interpret(nil)
}
