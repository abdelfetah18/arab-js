package main

import (
	"arab_js/internal/checker"
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/printer"
	"fmt"
)

func main() {
	lexer := compiler.NewLexer("متغير عدد : عدد = '100'؛")
	parser := compiler.NewParser(lexer, false)

	program := parser.Parse()

	checker := checker.NewChecker(program)
	checker.Check()

	if len(checker.Errors) > 0 {
		for _, msg := range checker.Errors {
			fmt.Println(msg)
		}
		return
	}

	_printer := printer.NewPrinter()
	_printer.Write(program)

	fmt.Printf("[ output ] ==============================\n%s\n==============================\n", _printer.Writer.Output)
}
