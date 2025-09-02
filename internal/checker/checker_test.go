package checker

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler"
	"testing"
)

func TestCheckVariableDeclaration(t *testing.T) {
	t.Run("should report error for wrong type", func(t *testing.T) {
		input := "متغير مرحبا: نص = 100؛"

		parser := compiler.NewParser(compiler.NewLexer(input), false)
		program := parser.Parse()
		_binder := binder.NewBinder(program)
		_binder.Bind()
		_checker := NewChecker(program)
		_checker.Check()

		if len(_checker.Errors) == 0 {
			t.Error("should detect error")
		}
	})

	t.Run("should not report error for correct type", func(t *testing.T) {
		input := "متغير رسالة: نص = 'مرحبا'؛"

		parser := compiler.NewParser(compiler.NewLexer(input), false)
		program := parser.Parse()
		_binder := binder.NewBinder(program)
		_binder.Bind()
		_checker := NewChecker(program)
		_checker.Check()

		if len(_checker.Errors) > 0 {
			t.Error("should not detect errors")
		}
	})
}
