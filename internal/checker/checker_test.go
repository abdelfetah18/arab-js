package checker

import (
	"arab_js/internal/binder"
	"arab_js/internal/compiler/ast"
	"arab_js/internal/compiler/parser"
	"testing"
)

type ProgramStub struct {
	sourceFiles []*ast.SourceFile
}

func (p ProgramStub) SourceFiles() []*ast.SourceFile { return p.sourceFiles }
func (p *ProgramStub) BindSourceFiles() {
	for _, sourceFile := range p.sourceFiles {
		binder.BindSourceFile(sourceFile)
	}
}
func (p *ProgramStub) CheckSourceFiles() *binder.NameResolver {
	c := NewChecker(p)
	c.Check()
	return c.NameResolver
}

func TestCheckVariableDeclaration(t *testing.T) {
	t.Run("should report error for wrong type", func(t *testing.T) {
		input := "متغير مرحبا: نص = 100؛"

		sourceFile := parser.ParseSourceFile(input)
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect error")
		}
	})

	t.Run("should not report error for correct type", func(t *testing.T) {
		input := "متغير رسالة: نص = 'مرحبا'؛"

		sourceFile := parser.ParseSourceFile(input)
		binder.NewBinder(sourceFile).Bind()
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) > 0 {
			t.Error("should not detect errors")
		}
	})
}

func TestCheckAssignmentExpression(t *testing.T) {
	t.Run("should report error for wrong type", func(t *testing.T) {
		input := "متغير مرحبا: نص = 'مرحبا'؛ مرحبا = 100؛"

		sourceFile := parser.ParseSourceFile(input)
		binder.NewBinder(sourceFile).Bind()
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect error")
		}
	})

	t.Run("should not report error for correct type", func(t *testing.T) {
		input := "متغير مرحبا: نص = 'مرحبا'؛ مرحبا = 'مرحبا بك'؛"

		sourceFile := parser.ParseSourceFile(input)
		binder.NewBinder(sourceFile).Bind()
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) > 0 {
			t.Error("should not detect errors")
		}
	})

	t.Run("should error on wrong type in object", func(t *testing.T) {
		input := "متغير شخص: { اسم: نص } = { اسم: \"شخص\" }؛\nشخص.اسم = 100؛"

		sourceFile := parser.ParseSourceFile(input)
		binder.NewBinder(sourceFile).Bind()
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect error")
		}
	})

	t.Run("should allow correct type in object", func(t *testing.T) {
		input := "متغير شخص: { اسم: نص } = { اسم: \"شخص\" }؛\nشخص.اسم = \"شخص\"؛"

		sourceFile := parser.ParseSourceFile(input)
		binder.BindSourceFile(sourceFile)
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) > 0 {
			t.Error("should not detect errors")
		}
	})
}

func TestCheckCallExpression(t *testing.T) {
	t.Run("should report an error when the type is not callable", func(t *testing.T) {
		input := "متغير مرحبا: نص = 'مرحبا'؛ مرحبا()؛"

		sourceFile := parser.ParseSourceFile(input)
		binder.NewBinder(sourceFile).Bind()
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect error")
		}
	})

	t.Run("should report an error when the number of parameters is incorrect", func(t *testing.T) {
		input := "دالة مرحبا(رسالة: نص) { } مرحبا()؛"

		sourceFile := parser.ParseSourceFile(input)
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect errors")
		}
	})

	t.Run("should report an error when the parameter type is incorrect", func(t *testing.T) {
		input := "دالة مرحبا(رسالة: نص) { } مرحبا(1)؛"

		sourceFile := parser.ParseSourceFile(input)
		binder.BindSourceFile(sourceFile)
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect errors")
		}
	})

	t.Run("should not report an error when the call is correct", func(t *testing.T) {
		input := "دالة مرحبا(رسالة: نص) { } مرحبا('مرحبا بك')؛"

		sourceFile := parser.ParseSourceFile(input)
		binder.BindSourceFile(sourceFile)
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) > 0 {
			t.Logf("_checker.Diagnostics='%v'", _checker.Diagnostics[0])
			t.Error("should not detect errors")
		}
	})
}

func TestCheckObjectExpression(t *testing.T) {
	t.Run("should report error for wrong type", func(t *testing.T) {
		input := "متغير شخص: { اسم: نص } = { اسم: 100 }؛"

		sourceFile := parser.ParseSourceFile(input)
		binder.NewBinder(sourceFile).Bind()
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect error")
		}
	})

	t.Run("should not report error for correct type", func(t *testing.T) {
		input := "متغير شخص: { اسم: نص } = { اسم: \"شخص\" }؛"

		sourceFile := parser.ParseSourceFile(input)
		binder.BindSourceFile(sourceFile)
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) > 0 {
			t.Error("should not detect errors")
		}
	})
}

func TestCheckMemberExpression(t *testing.T) {
	t.Run("should report error when property does not exist", func(t *testing.T) {
		input := "متغير شخص: { اسم: نص } = { اسم: \"شخص\" }؛\nشخص.رقم = 100؛"

		sourceFile := parser.ParseSourceFile(input)
		binder.BindSourceFile(sourceFile)
		_checker := NewChecker(&ProgramStub{sourceFiles: []*ast.SourceFile{sourceFile}})
		_checker.Check()

		if len(_checker.Diagnostics) == 0 {
			t.Error("should detect errors")
		}
	})
}
