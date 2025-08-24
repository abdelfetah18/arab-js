package main

import (
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/printer"
	"fmt"
)

func main() {
	lexer := compiler.NewLexer("إذا (100) { متغير عدد = 100؛ }")
	parser := compiler.NewParser(lexer, false)

	program := parser.Parse()

	_printer := printer.NewPrinter()
	_printer.Write(program)

	fmt.Printf("[ output ] ==============================\n%s\n==============================\n", _printer.Writer.Output)
}
