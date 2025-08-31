package transformer

import (
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/ast"
	"reflect"
	"testing"
)

func TestOriginalName(t *testing.T) {
	t.Run("should update identifiers to original name", func(t *testing.T) {
		input := "// @الاسم_الأصلي(\"console\")\nتصريح متغير وحدة_التحكم؛\nوحدة_التحكم؛"

		identifier := ast.NewIdentifier("console", nil)
		s := "console"
		identifier.OriginalName = &s

		expected := ast.NewProgram([]*ast.Node{
			ast.NewVariableDeclaration(
				identifier,
				nil,
				true,
			).ToNode(),
			ast.NewExpressionStatement(ast.NewIdentifier("console", nil).ToNode()).ToNode(),
		}, []*ast.Directive{})

		parser := compiler.NewParser(compiler.NewLexer(input), false)
		program := parser.Parse()
		symbolTable := ast.BuildSymbolTable(program)

		transformer := NewTransformer(program, symbolTable)
		transformer.Transform()

		if !reflect.DeepEqual(program, expected) {
			t.Error("AST structures are not equal")
		}
	})
}
