package printer

import (
	"arab_js/internal/compiler"
	"testing"
)

func TestVariableDeclaration(t *testing.T) {
	t.Run("should parse a VariableDeclaration", func(t *testing.T) {
		input := "متغير عدد = 100؛"
		parser := compiler.NewParser(compiler.NewLexer(input), false)
		program := parser.Parse()

		printer := NewPrinter()
		printer.Write(program)

		expected := "let عدد = 100;"

		if printer.Writer.Output != expected {
			t.Errorf("\nExpected %s, got %s\n", expected, printer.Writer.Output)
		}
	})

	t.Run("should parse a VariableDeclaration inside a BlockStatement", func(t *testing.T) {
		input := "{ متغير عدد = 100؛ }"
		parser := compiler.NewParser(compiler.NewLexer(input), false)
		program := parser.Parse()

		printer := NewPrinter()
		printer.Write(program)

		expected := "{\n  let عدد = 100;\n}"

		if printer.Writer.Output != expected {
			t.Errorf("\nExpected:\n%s\nGot:\n%s\n", expected, printer.Writer.Output)
		}
	})

	t.Run("should throw on VariableDeclaration missing semicolon", func(t *testing.T) {
		input := "متغير عدد = 100"
		parser := compiler.NewParser(compiler.NewLexer(input), false)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("\nExpected panic for missing semicolon, but no panic occurred")
			}
		}()

		program := parser.Parse()
		printer := NewPrinter()
		printer.Write(program)
	})

	t.Run("should throw on VariableDeclaration with invalid identifier", func(t *testing.T) {
		input := "متغير عد1د = 100؛"
		parser := compiler.NewParser(compiler.NewLexer(input), false)

		defer func() {
			if r := recover(); r == nil {
				t.Errorf("\nExpected panic for invalid identifier, but no panic occurred")
			}
		}()

		program := parser.Parse()
		printer := NewPrinter()
		printer.Write(program)
	})
}

func TestStringLiteral(t *testing.T) {
	t.Run("should parse a double-quoted StringLiteral", func(t *testing.T) {
		input := "متغير نص = \"أهلا\"؛"
		parser := compiler.NewParser(compiler.NewLexer(input), false)
		program := parser.Parse()

		printer := NewPrinter()
		printer.Write(program)

		expected := "let نص = \"أهلا\";"

		if printer.Writer.Output != expected {
			t.Errorf("Expected %s, got %s\n", expected, printer.Writer.Output)
		}
	})

	t.Run("should parse a single-quoted StringLiteral", func(t *testing.T) {
		input := "متغير نص = 'أهلا'؛"
		parser := compiler.NewParser(compiler.NewLexer(input), false)
		program := parser.Parse()

		printer := NewPrinter()
		printer.Write(program)

		expected := "let نص = \"أهلا\";"

		if printer.Writer.Output != expected {
			t.Errorf("Expected %s, got %s\n", expected, printer.Writer.Output)
		}
	})

	t.Run("should parse a StringLiteral inside a BlockStatement", func(t *testing.T) {
		input := "{ اطبع(\"مرحبا\")؛ }"
		parser := compiler.NewParser(compiler.NewLexer(input), false)
		program := parser.Parse()

		printer := NewPrinter()
		printer.Write(program)

		expected := "{\n  اطبع(\"مرحبا\");\n}"

		if printer.Writer.Output != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s\n", expected, printer.Writer.Output)
		}
	})
}
