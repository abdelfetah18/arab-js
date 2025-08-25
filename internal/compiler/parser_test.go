package compiler

import (
	"arab_js/internal/compiler/ast"
	"reflect"
	"testing"
)

func TestTInterfaceDeclaration(t *testing.T) {
	t.Run("should parse interface declaration", func(t *testing.T) {
		input := "واجهة مستخدم { الاسم: نص }"

		expected := ast.NewProgram([]*ast.Node{
			ast.NewTInterfaceDeclaration(
				ast.NewIdentifier("مستخدم"),
				ast.NewTInterfaceBody([]*ast.Node{
					ast.NewTPropertySignature(
						ast.NewIdentifier("الاسم").ToNode(),
						ast.NewTTypeAnnotation(
							ast.NewTStringKeyword().ToNode(),
						),
					).ToNode(),
				})).ToNode(),
		}, []*ast.Directive{})

		parser := NewParser(NewLexer(input), false)
		program := parser.Parse()

		if !reflect.DeepEqual(program, expected) {
			t.Error("AST structures are not equal")
		}
	})

	t.Run("should parse interface with multiple properties", func(t *testing.T) {
		input := "واجهة بيانات { النص: نص, الرقم: عدد, الحالة: قيمة_منطقية }"

		expected := ast.NewProgram([]*ast.Node{
			ast.NewTInterfaceDeclaration(
				ast.NewIdentifier("بيانات"),
				ast.NewTInterfaceBody([]*ast.Node{
					ast.NewTPropertySignature(
						ast.NewIdentifier("النص").ToNode(),
						ast.NewTTypeAnnotation(
							ast.NewTStringKeyword().ToNode(),
						),
					).ToNode(),
					ast.NewTPropertySignature(
						ast.NewIdentifier("الرقم").ToNode(),
						ast.NewTTypeAnnotation(
							ast.NewTNumberKeyword().ToNode(),
						),
					).ToNode(),
					ast.NewTPropertySignature(
						ast.NewIdentifier("الحالة").ToNode(),
						ast.NewTTypeAnnotation(
							ast.NewTBooleanKeyword().ToNode(),
						),
					).ToNode(),
				})).ToNode(),
		}, []*ast.Directive{})

		parser := NewParser(NewLexer(input), false)
		program := parser.Parse()

		if !reflect.DeepEqual(program, expected) {
			t.Error("AST structures are not equal")
		}
	})
}
