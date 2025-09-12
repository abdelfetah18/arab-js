package checker

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/ast"
	"testing"
)

func TestCheckVariableDeclaration(t *testing.T) {
	t.Run("should report error for wrong type", func(t *testing.T) {
		input := "متغير مرحبا: نص = 100؛"

		sourceFile := compiler.ParseSourceFile(input)
		binder.NewBinder(sourceFile).Bind()
		_checker := NewChecker(compiler.NewProgram([]*ast.SourceFile{sourceFile}))
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect error")
		}
	})

	t.Run("should not report error for correct type", func(t *testing.T) {
		input := "متغير رسالة: نص = 'مرحبا'؛"

		sourceFile := compiler.ParseSourceFile(input)
		binder.NewBinder(sourceFile).Bind()
		_checker := NewChecker(compiler.NewProgram([]*ast.SourceFile{sourceFile}))
		_checker.Check()

		if len(_checker.Diagnostics) > 0 {
			t.Error("should not detect errors")
		}
	})
}

func TestCheckAssignmentExpression(t *testing.T) {
	t.Run("should report error for wrong type", func(t *testing.T) {
		input := "متغير مرحبا: نص = 'مرحبا'؛ مرحبا = 100؛"

		sourceFile := compiler.ParseSourceFile(input)
		binder.NewBinder(sourceFile).Bind()
		_checker := NewChecker(compiler.NewProgram([]*ast.SourceFile{sourceFile}))
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect error")
		}
	})

	t.Run("should not report error for correct type", func(t *testing.T) {
		input := "متغير مرحبا: نص = 'مرحبا'؛ مرحبا = 'مرحبا بك'؛"

		sourceFile := compiler.ParseSourceFile(input)
		binder.NewBinder(sourceFile).Bind()
		_checker := NewChecker(compiler.NewProgram([]*ast.SourceFile{sourceFile}))
		_checker.Check()

		if len(_checker.Diagnostics) > 0 {
			t.Error("should not detect errors")
		}
	})
}
