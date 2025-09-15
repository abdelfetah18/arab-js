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

func TestCheckCallExpression(t *testing.T) {
	t.Run("should report an error when the type is not callable", func(t *testing.T) {
		input := "متغير مرحبا: نص = 'مرحبا'؛ مرحبا()؛"

		sourceFile := compiler.ParseSourceFile(input)
		binder.NewBinder(sourceFile).Bind()
		_checker := NewChecker(compiler.NewProgram([]*ast.SourceFile{sourceFile}))
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect error")
		}
	})

	t.Run("should report an error when the number of parameters is incorrect", func(t *testing.T) {
		input := "دالة مرحبا(رسالة: نص) { } مرحبا()؛"

		sourceFile := compiler.ParseSourceFile(input)
		binder.BindSourceFile(sourceFile)
		_checker := NewChecker(compiler.NewProgram([]*ast.SourceFile{sourceFile}))
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect errors")
		}
	})

	t.Run("should report an error when the parameter type is incorrect", func(t *testing.T) {
		input := "دالة مرحبا(رسالة: نص) { } مرحبا(1)؛"

		sourceFile := compiler.ParseSourceFile(input)
		binder.BindSourceFile(sourceFile)
		_checker := NewChecker(compiler.NewProgram([]*ast.SourceFile{sourceFile}))
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect errors")
		}
	})

	t.Run("should not report an error when the call is correct", func(t *testing.T) {
		input := "دالة مرحبا(رسالة: نص) { } مرحبا('مرحبا بك')؛"

		sourceFile := compiler.ParseSourceFile(input)
		binder.BindSourceFile(sourceFile)
		_checker := NewChecker(compiler.NewProgram([]*ast.SourceFile{sourceFile}))
		_checker.Check()

		if len(_checker.Diagnostics) > 0 {
			t.Error("should not detect errors")
		}
	})
}
