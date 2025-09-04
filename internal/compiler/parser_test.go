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

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
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

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
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

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
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
				[]*ast.Node{
					ast.NewIdentifier("أ", ast.NewTTypeAnnotation(ast.NewTNumberKeyword().ToNode())).ToNode(),
					ast.NewIdentifier("ب", ast.NewTTypeAnnotation(ast.NewTNumberKeyword().ToNode())).ToNode(),
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

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
			t.Error("AST structures are not equal")
		}
	})

	t.Run("should parse function declaration with rest element", func(t *testing.T) {
		input := "دالة جمع (...أعداد: عدد[]) : عدد {}"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewFunctionDeclaration(
				ast.NewIdentifier("جمع", nil),
				[]*ast.Node{
					ast.NewRestElement(
						ast.NewIdentifier("أعداد", nil),
						ast.NewTTypeAnnotation(
							ast.NewTArrayType(ast.NewTNumberKeyword().ToNode()).ToNode(),
						),
					).ToNode(),
				},
				ast.NewBlockStatement([]*ast.Node{}),
				ast.NewTTypeAnnotation(ast.NewTNumberKeyword().ToNode()),
			).ToNode(),
		}, []*ast.Directive{})

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
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
								[]*ast.Node{
									ast.NewIdentifier(
										"اسم",
										ast.NewTTypeAnnotation(
											ast.NewTStringKeyword().ToNode()),
									).ToNode(),
								},
								ast.NewTTypeAnnotation(ast.NewTStringKeyword().ToNode()),
							).ToNode(),
						),
					).ToNode(),
				})).ToNode(),
		}, []*ast.Directive{})

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
			t.Error("AST structures are not equal")
		}
	})

	t.Run("should parse function type with rest param in interface", func(t *testing.T) {
		input := "واجهة حاسبة { جمع: (...أعداد: عدد[]) => عدد }"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewTInterfaceDeclaration(
				ast.NewIdentifier("حاسبة", nil),
				ast.NewTInterfaceBody([]*ast.Node{
					ast.NewTPropertySignature(
						ast.NewIdentifier("جمع", nil).ToNode(),
						ast.NewTTypeAnnotation(
							ast.NewTFunctionType(
								[]*ast.Node{
									ast.NewRestElement(
										ast.NewIdentifier("أعداد", nil),
										ast.NewTTypeAnnotation(
											ast.NewTArrayType(ast.NewTNumberKeyword().ToNode()).ToNode(),
										),
									).ToNode(),
								},
								ast.NewTTypeAnnotation(ast.NewTNumberKeyword().ToNode()),
							).ToNode(),
						),
					).ToNode(),
				})).ToNode(),
		}, []*ast.Directive{})

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
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

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
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

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
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

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
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

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
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

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
			t.Error("AST structures are not equal")
		}
	})
}
func TestTypeKeywords(t *testing.T) {
	t.Run("should parse type any", func(t *testing.T) {
		input := "متغير رقم : أي_نوع = 100؛"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewVariableDeclaration(
				ast.NewIdentifier(
					"رقم",
					ast.NewTTypeAnnotation(ast.NewTAnyKeyword().ToNode()),
				),
				ast.NewInitializer(ast.NewDecimalLiteral("100").ToNode()),
				false,
			).ToNode(),
		}, []*ast.Directive{})

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
			t.Error("AST structures are not equal")
		}
	})

	t.Run("should parse type number", func(t *testing.T) {
		input := "متغير رقم : عدد = 100؛"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewVariableDeclaration(
				ast.NewIdentifier(
					"رقم",
					ast.NewTTypeAnnotation(ast.NewTNumberKeyword().ToNode()),
				),
				ast.NewInitializer(ast.NewDecimalLiteral("100").ToNode()),
				false,
			).ToNode(),
		}, []*ast.Directive{})

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
			t.Error("AST structures are not equal")
		}
	})

	t.Run("should parse type string", func(t *testing.T) {
		input := "متغير جملة : نص = 'مرحبا'؛"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewVariableDeclaration(
				ast.NewIdentifier(
					"جملة",
					ast.NewTTypeAnnotation(ast.NewTStringKeyword().ToNode()),
				),
				ast.NewInitializer(ast.NewStringLiteral("مرحبا").ToNode()),
				false,
			).ToNode(),
		}, []*ast.Directive{})

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
			t.Error("AST structures are not equal")
		}
	})

	t.Run("should parse type boolean", func(t *testing.T) {
		input := "متغير قيمة_منطقية : قيمة_منطقية = صحيح؛"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewVariableDeclaration(
				ast.NewIdentifier(
					"قيمة_منطقية",
					ast.NewTTypeAnnotation(ast.NewTBooleanKeyword().ToNode()),
				),
				ast.NewInitializer(ast.NewBooleanLiteral(true).ToNode()),
				false,
			).ToNode(),
		}, []*ast.Directive{})

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
			t.Error("AST structures are not equal")
		}
	})

	t.Run("should parse type array", func(t *testing.T) {
		input := "متغير مصفوفة_أعداد : عدد[] = []؛"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewVariableDeclaration(
				ast.NewIdentifier(
					"مصفوفة_أعداد",
					ast.NewTTypeAnnotation(ast.NewTArrayType(ast.NewTNumberKeyword().ToNode()).ToNode()),
				),
				ast.NewInitializer(ast.NewArrayExpression([]*ast.Node{}).ToNode()),
				false,
			).ToNode(),
		}, []*ast.Directive{})

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
			t.Error("AST structures are not equal")
		}
	})
}

func TestUpdateExpression(t *testing.T) {
	t.Run("should parse update expression", func(t *testing.T) {
		input := "أ++؛"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewExpressionStatement(
				ast.NewUpdateExpression(
					ast.PLUS_PLUS,
					ast.NewIdentifier("أ", nil).ToNode(),
					false,
				).ToNode(),
			).ToNode(),
		}, []*ast.Directive{})

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
			t.Error("AST structures are not equal")
		}
	})
}

func TestForStatement(t *testing.T) {
	t.Run("should parse for statement", func(t *testing.T) {
		input := "من_أجل (متغير أ = 0؛ أ <= 10؛ أ++) { أ؛ }"

		expected := ast.NewSourceFile([]*ast.Node{
			ast.NewForStatement(
				ast.NewVariableDeclaration(ast.NewIdentifier("أ", nil), ast.NewInitializer(ast.NewDecimalLiteral("0").ToNode()), false).ToNode(),
				ast.NewBinaryExpression(ast.LEFT_ARROW_EQUAL, ast.NewIdentifier("أ", nil).ToNode(), ast.NewDecimalLiteral("10").ToNode()).ToNode(),
				ast.NewUpdateExpression(ast.PLUS_PLUS, ast.NewIdentifier("أ", nil).ToNode(), false).ToNode(),
				ast.NewBlockStatement([]*ast.Node{
					ast.NewExpressionStatement(ast.NewIdentifier("أ", nil).ToNode()).ToNode(),
				}).ToNode(),
			).ToNode(),
		}, []*ast.Directive{})

		sourceFile := ParseSourceFile(input)

		if !reflect.DeepEqual(sourceFile, expected) {
			t.Error("AST structures are not equal")
		}
	})
}
