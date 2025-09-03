package compiler

import (
	"arab_js/internal/compiler/ast"
	"reflect"
	"testing"
)

func TestTInterfaceDeclaration(t *testing.T) {
	t.Run("should parse interface declaration", func(t *testing.T) {
		input := "واجهة مستخدم { الاسم: نص }"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewTInterfaceDeclaration(
				ast.NewIdentifier("مستخدم", nil),
				ast.NewTInterfaceBody([]*ast.Node{
					ast.NewTPropertySignature(
						ast.NewIdentifier("الاسم", nil).ToNode(),
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

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewTInterfaceDeclaration(
				ast.NewIdentifier("بيانات", nil),
				ast.NewTInterfaceBody([]*ast.Node{
					ast.NewTPropertySignature(
						ast.NewIdentifier("النص", nil).ToNode(),
						ast.NewTTypeAnnotation(
							ast.NewTStringKeyword().ToNode(),
						),
					).ToNode(),
					ast.NewTPropertySignature(
						ast.NewIdentifier("الرقم", nil).ToNode(),
						ast.NewTTypeAnnotation(
							ast.NewTNumberKeyword().ToNode(),
						),
					).ToNode(),
					ast.NewTPropertySignature(
						ast.NewIdentifier("الحالة", nil).ToNode(),
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

	t.Run("should parse interface declaration with a property that has an original name", func(t *testing.T) {
		input := "واجهة مستخدم {\n// @الاسم_الأصلي(\"name\")\nالاسم: نص\n}"

		identifier := ast.NewIdentifier("الاسم", nil)
		s := "name"
		identifier.OriginalName = &s
		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewTInterfaceDeclaration(
				ast.NewIdentifier("مستخدم", nil),
				ast.NewTInterfaceBody([]*ast.Node{
					ast.NewTPropertySignature(
						identifier.ToNode(),
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
}

func TestFunctionDeclaration(t *testing.T) {
	t.Run("should parse typed function", func(t *testing.T) {
		input := "دالة جمع (أ: عدد, ب: عدد) : عدد { إرجاع أ + ب؛ }"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewFunctionDeclaration(
				ast.NewIdentifier("جمع", nil),
				[]*ast.Identifier{
					ast.NewIdentifier("أ", ast.NewTTypeAnnotation(ast.NewTNumberKeyword().ToNode())),
					ast.NewIdentifier("ب", ast.NewTTypeAnnotation(ast.NewTNumberKeyword().ToNode())),
				},
				ast.NewBlockStatement([]*ast.Node{
					ast.NewReturnStatement(
						ast.NewBinaryExpression(
							"+",
							ast.NewIdentifier("أ", nil).ToNode(),
							ast.NewIdentifier("ب", nil).ToNode(),
						).ToNode(),
					).ToNode(),
				}),
				ast.NewTTypeAnnotation(ast.NewTNumberKeyword().ToNode()),
			).ToNode(),
		}, []*ast.Directive{})

		parser := NewParser(NewLexer(input), false)
		program := parser.Parse()

		if !reflect.DeepEqual(program, expected) {
			t.Error("AST structures are not equal")
		}
	})
}

func TestTFunctionType(t *testing.T) {
	t.Run("should parse interface declaration with function type", func(t *testing.T) {
		input := "واجهة مستخدم { جلب_بيانات_المستخدم: (اسم:نص) => نص }"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewTInterfaceDeclaration(
				ast.NewIdentifier("مستخدم", nil),
				ast.NewTInterfaceBody([]*ast.Node{
					ast.NewTPropertySignature(
						ast.NewIdentifier("جلب_بيانات_المستخدم", nil).ToNode(),
						ast.NewTTypeAnnotation(
							ast.NewTFunctionType(
								[]*ast.Identifier{
									ast.NewIdentifier(
										"اسم",
										ast.NewTTypeAnnotation(
											ast.NewTStringKeyword().ToNode()),
									),
								},
								ast.NewTTypeAnnotation(ast.NewTStringKeyword().ToNode()),
							).ToNode(),
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

func TestTTypeAliasDeclaration(t *testing.T) {
	t.Run("should parse type alias declaration", func(t *testing.T) {
		input := "نوع الاسم = نص"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewTTypeAliasDeclaration(
				ast.NewIdentifier("الاسم", nil),
				ast.NewTTypeAnnotation(ast.NewTStringKeyword().ToNode()),
			).ToNode(),
		}, []*ast.Directive{})

		parser := NewParser(NewLexer(input), false)
		program := parser.Parse()

		if !reflect.DeepEqual(program, expected) {
			t.Error("AST structures are not equal")
		}
	})

}

func TestTTypeLiteral(t *testing.T) {
	t.Run("should parse type literal", func(t *testing.T) {
		input := "نوع مستخدم = { الاسم: نص }"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewTTypeAliasDeclaration(
				ast.NewIdentifier("مستخدم", nil),
				ast.NewTTypeAnnotation(
					ast.NewTTypeLiteral([]*ast.Node{
						ast.NewTPropertySignature(
							ast.NewIdentifier("الاسم", nil).ToNode(),
							ast.NewTTypeAnnotation(
								ast.NewTStringKeyword().ToNode(),
							),
						).ToNode(),
					}).ToNode(),
				),
			).ToNode(),
		}, []*ast.Directive{})

		parser := NewParser(NewLexer(input), false)
		program := parser.Parse()

		if !reflect.DeepEqual(program, expected) {
			t.Error("AST structures are not equal")
		}
	})

	t.Run("should parse type literal with multiple properties", func(t *testing.T) {
		input := "نوع بيانات = { النص: نص, الرقم: عدد, الحالة: قيمة_منطقية }"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewTTypeAliasDeclaration(
				ast.NewIdentifier("بيانات", nil),
				ast.NewTTypeAnnotation(
					ast.NewTTypeLiteral([]*ast.Node{
						ast.NewTPropertySignature(
							ast.NewIdentifier("النص", nil).ToNode(),
							ast.NewTTypeAnnotation(
								ast.NewTStringKeyword().ToNode(),
							),
						).ToNode(),
						ast.NewTPropertySignature(
							ast.NewIdentifier("الرقم", nil).ToNode(),
							ast.NewTTypeAnnotation(
								ast.NewTNumberKeyword().ToNode(),
							),
						).ToNode(),
						ast.NewTPropertySignature(
							ast.NewIdentifier("الحالة", nil).ToNode(),
							ast.NewTTypeAnnotation(
								ast.NewTBooleanKeyword().ToNode(),
							),
						).ToNode(),
					}).ToNode(),
				),
			).ToNode(),
		}, []*ast.Directive{})

		parser := NewParser(NewLexer(input), false)
		program := parser.Parse()

		if !reflect.DeepEqual(program, expected) {
			t.Error("AST structures are not equal")
		}
	})
}

func TestVariableDeclaration(t *testing.T) {
	t.Run("should parse variable declaration with declare keyword", func(t *testing.T) {
		input := "تصريح متغير رقم: عدد؛"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewVariableDeclaration(
				ast.NewIdentifier(
					"رقم",
					ast.NewTTypeAnnotation(ast.NewTNumberKeyword().ToNode()),
				),
				nil,
				true,
			).ToNode(),
		}, []*ast.Directive{})

		parser := NewParser(NewLexer(input), false)
		program := parser.Parse()

		if !reflect.DeepEqual(program, expected) {
			t.Error("AST structures are not equal")
		}
	})

	t.Run("should parse variable declaration with declare keyword and has original name", func(t *testing.T) {
		input := "// @الاسم_الأصلي(\"num\")\nتصريح متغير رقم: عدد؛"

		identifier := ast.NewIdentifier(
			"رقم",
			ast.NewTTypeAnnotation(ast.NewTNumberKeyword().ToNode()),
		)

		originalName := "num"
		identifier.OriginalName = &originalName

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewVariableDeclaration(
				identifier,
				nil,
				true,
			).ToNode(),
		}, []*ast.Directive{})

		parser := NewParser(NewLexer(input), false)
		program := parser.Parse()

		if !reflect.DeepEqual(program, expected) {
			t.Error("AST structures are not equal")
		}
	})
}
