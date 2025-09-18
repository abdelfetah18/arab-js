package printer

import (
	"arab_js/internal/compiler/lexer"
	"arab_js/internal/compiler/parser"
	"testing"
)

func TestVariableDeclaration(t *testing.T) {
	t.Run("should parse a VariableDeclaration", func(t *testing.T) {
		input := "متغير عدد = 100؛"
		sourceFile := parser.ParseSourceFile(input)
		output := WriteSourceFile(sourceFile)
		expected := "let عدد = 100;"

		if output != expected {
			t.Errorf("\nExpected %s, got %s\n", expected, output)
		}
	})

	t.Run("should parse a VariableDeclaration inside a BlockStatement", func(t *testing.T) {
		input := "{ متغير عدد = 100؛ }"
		sourceFile := parser.ParseSourceFile(input)
		output := WriteSourceFile(sourceFile)
		expected := "{\n  let عدد = 100;\n}"

		if output != expected {
			t.Errorf("\nExpected:\n%s\nGot:\n%s\n", expected, output)
		}
	})

	t.Run("should throw on VariableDeclaration missing semicolon", func(t *testing.T) {
		input := "متغير عدد = 100"

		parser := parser.NewParser(lexer.NewLexer(input))
		sourceFile := parser.Parse()
		WriteSourceFile(sourceFile)

		if len(parser.Diagnostics) == 0 {
			t.Errorf("\nExpected errors for missing semicolon")
		}
	})

	t.Run("should throw on VariableDeclaration with invalid identifier", func(t *testing.T) {
		input := "متغير عد1د = 100؛"

		parser := parser.NewParser(lexer.NewLexer(input))
		sourceFile := parser.Parse()
		WriteSourceFile(sourceFile)

		if len(parser.Diagnostics) == 0 {
			t.Errorf("\nExpected errors for invalid identifier")
		}
	})
}

func TestStringLiteral(t *testing.T) {
	t.Run("should parse a double-quoted StringLiteral", func(t *testing.T) {
		input := "متغير نص = \"أهلا\"؛"
		sourceFile := parser.ParseSourceFile(input)
		output := WriteSourceFile(sourceFile)
		expected := "let نص = \"أهلا\";"

		if output != expected {
			t.Errorf("Expected %s, got %s\n", expected, output)
		}
	})

	t.Run("should parse a single-quoted StringLiteral", func(t *testing.T) {
		input := "متغير نص = 'أهلا'؛"
		sourceFile := parser.ParseSourceFile(input)
		output := WriteSourceFile(sourceFile)
		expected := "let نص = \"أهلا\";"

		if output != expected {
			t.Errorf("Expected %s, got %s\n", expected, output)
		}
	})

	t.Run("should parse a StringLiteral inside a BlockStatement", func(t *testing.T) {
		input := "{ اطبع(\"مرحبا\")؛ }"
		sourceFile := parser.ParseSourceFile(input)
		output := WriteSourceFile(sourceFile)
		expected := "{\n  اطبع(\"مرحبا\");\n}"

		if output != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s\n", expected, output)
		}
	})
}

func TestBlockStatement(t *testing.T) {
	t.Run("should parse an empty BlockStatement", func(t *testing.T) {
		input := "{}"
		sourceFile := parser.ParseSourceFile(input)
		output := WriteSourceFile(sourceFile)
		expected := "{}"

		if output != expected {
			t.Errorf("Expected %s, got %s\n", expected, output)
		}
	})

	t.Run("should parse a BlockStatement with a single VariableDeclaration", func(t *testing.T) {
		input := "{ متغير عدد = 100؛ }"
		sourceFile := parser.ParseSourceFile(input)
		output := WriteSourceFile(sourceFile)
		expected := "{\n  let عدد = 100;\n}"

		if output != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s\n", expected, output)
		}
	})

	t.Run("should parse a BlockStatement with multiple statements", func(t *testing.T) {
		input := "{ متغير عدد = 100؛ متغير رقم = 1؛}"
		sourceFile := parser.ParseSourceFile(input)
		output := WriteSourceFile(sourceFile)
		expected := "{\n  let عدد = 100;\n  let رقم = 1;\n}"

		if output != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s\n", expected, output)
		}
	})

	t.Run("should parse a nested BlockStatement", func(t *testing.T) {
		input := "{ { متغير س = 5؛ } }"
		sourceFile := parser.ParseSourceFile(input)
		output := WriteSourceFile(sourceFile)
		expected := "{\n  {\n    let س = 5;\n  }\n}"

		if output != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s\n", expected, output)
		}
	})

	t.Run("should parse a BlockStatement with different statement types", func(t *testing.T) {
		input := "{ احسب(1,2)؛ متغير س = 10؛ }"
		sourceFile := parser.ParseSourceFile(input)
		output := WriteSourceFile(sourceFile)
		expected := "{\n  احسب(1, 2);\n  let س = 10;\n}"

		if output != expected {
			t.Errorf("Expected:\n%s\nGot:\n%s\n", expected, output)
		}
	})

	t.Run("should throw on BlockStatement with missing closing brace", func(t *testing.T) {
		input := "{ متغير عدد = 100؛ "

		parser := parser.NewParser(lexer.NewLexer(input))
		sourceFile := parser.Parse()
		WriteSourceFile(sourceFile)

		if len(parser.Diagnostics) == 0 {
			t.Errorf("\nExpected errors for missing closing brace")
		}
	})
}

func TestBinaryExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Bitwise OR expression",
			input:    "متغير عدد = 100 | 1؛",
			expected: "let عدد = 100 | 1;",
		},
		{
			name:     "Multiple Bitwise OR expressions",
			input:    "متغير عدد = 100 | 1 | 2؛",
			expected: "let عدد = 100 | 1 | 2;",
		},
		{
			name:     "Bitwise XOR expression",
			input:    "متغير عدد = 100 ^ 1؛",
			expected: "let عدد = 100 ^ 1;",
		},
		{
			name:     "Multiple Bitwise XOR expressions",
			input:    "متغير عدد = 100 ^ 1 ^ 2؛",
			expected: "let عدد = 100 ^ 1 ^ 2;",
		},
		{
			name:     "Bitwise AND expression",
			input:    "متغير عدد = 100 & 1؛",
			expected: "let عدد = 100 & 1;",
		},
		{
			name:     "Multiple Bitwise AND expressions",
			input:    "متغير عدد = 100 & 1 & 2؛",
			expected: "let عدد = 100 & 1 & 2;",
		},
		{
			name:     "Equality expressions",
			input:    "متغير يساوي = 100 == 1؛\nمتغير يساوي_يساوي = 100 === 1؛\nمتغير لا_يساوي = 100 != 1؛\nمتغير لا_يساوي_يساوي = 100 !== 1؛",
			expected: "let يساوي = 100 == 1;\nlet يساوي_يساوي = 100 === 1;\nlet لا_يساوي = 100 != 1;\nlet لا_يساوي_يساوي = 100 !== 1;",
		},
		{
			name:     "Multiple Equality expressions",
			input:    "متغير عدد = 100 == 1 != 2 === 3 !== 300؛",
			expected: "let عدد = 100 == 1 != 2 === 3 !== 300;",
		},
		{
			name:     "Relational expressions",
			input:    "متغير أكبر_من = 100 > 1؛\nمتغير أصغر_من = 100 < 1؛\nمتغير أكبر_أو_يساوي = 100 >= 1؛\nمتغير أصغر_أو_يساوي = 100 <= 1؛",
			expected: "let أكبر_من = 100 > 1;\nlet أصغر_من = 100 < 1;\nlet أكبر_أو_يساوي = 100 >= 1;\nlet أصغر_أو_يساوي = 100 <= 1;",
		},
		{
			name:     "Multiple Relational expressions",
			input:    "متغير عدد = 100 == 1 != 2 > 3 <= 300؛",
			expected: "let عدد = 100 == 1 != 2 > 3 <= 300;",
		},
		{
			name:     "Shift expressions",
			input:    "متغير أ = 100 >> 1؛\nمتغير ب = 100 << 1؛\nمتغير ت = 100 >>> 1؛",
			expected: "let أ = 100 >> 1;\nlet ب = 100 << 1;\nlet ت = 100 >>> 1;",
		},
		{
			name:     "Multiple Shift expressions",
			input:    "متغير عدد = 100 == 1 >> 2 >>> 3 << 300؛",
			expected: "let عدد = 100 == 1 >> 2 >>> 3 << 300;",
		},
		{
			name:     "Additive expressions",
			input:    "متغير جمع = 100 + 1؛\nمتغير طرح = 100 - 1؛",
			expected: "let جمع = 100 + 1;\nlet طرح = 100 - 1;",
		},
		{
			name:     "Multiple Additive expressions",
			input:    "متغير عملية_حسابية = 100 == 1 - 2 - 3 + 300؛",
			expected: "let عملية_حسابية = 100 == 1 - 2 - 3 + 300;",
		},
		{
			name:     "Multiplicative expressions",
			input:    "متغير ضرب = 100 * 1؛\nمتغير قسمة = 100 / 1؛\nمتغير باقي_القسمة = 100 % 1؛",
			expected: "let ضرب = 100 * 1;\nlet قسمة = 100 / 1;\nlet باقي_القسمة = 100 % 1;",
		},
		{
			name:     "Multiple Multiplicative expressions",
			input:    "متغير عملية_حسابية = 100 == 1 % 2 / 3 * 300؛",
			expected: "let عملية_حسابية = 100 == 1 % 2 / 3 * 300;",
		},
		{
			name:     "Exponentiation expression",
			input:    "متغير عملية_حسابية = 100 ** 1؛",
			expected: "let عملية_حسابية = 100 ** 1;",
		},
		{
			name:     "Multiple Exponentiation expressions",
			input:    "متغير عملية_حسابية = 1 ** 2 ** 3 ** 4؛",
			expected: "let عملية_حسابية = 1 ** 2 ** 3 ** 4;",
		},
		{
			name:     "Random Binary expressions",
			input:    "متغير عملية_حسابية = 1 | 2 ^ 3 & 4 == 5 === 6 != 7 !== 8 < 9 <= 10 > 11 >= 12 << 13 >> 14 >>> 15 + 16 - 17 * 18 / 19 % 20 ** 21؛",
			expected: "let عملية_حسابية = 1 | 2 ^ 3 & 4 == 5 === 6 != 7 !== 8 < 9 <= 10 > 11 >= 12 << 13 >> 14 >>> 15 + 16 - 17 * 18 / 19 % 20 ** 21;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(tt.input)
			output := WriteSourceFile(sourceFile)

			if output != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s\n", tt.expected, output)
			}
		})
	}
}

func TestArrayExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "should parse an empty array",
			input:    "[]؛",
			expected: "[];",
		},
		{
			name:     "should parse an array",
			input:    "[1, 'مرحبا', صحيح]؛",
			expected: `[1, "مرحبا", true];`,
		},
		{
			name:     "should parse a spread element in array",
			input:    "[1, 'مرحبا', صحيح, ...[1, 'مرحبا', صحيح]]؛",
			expected: `[1, "مرحبا", true, ...[1, "مرحبا", true]];`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(tt.input)
			output := WriteSourceFile(sourceFile)

			if output != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s\n", tt.expected, output)
			}
		})
	}
}

func TestObjectExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "should parse an empty object",
			input:    "متغير شيئ = {}؛",
			expected: "let شيئ = {};",
		},
		{
			name:  "should parse an object",
			input: "متغير شيئ = { رقم: 1, قيمة_منطقية: صحيح, نص: 'مرحبا', 1:'رقم واحد', 'رقم': 1 }؛",
			expected: `let شيئ = {
  رقم: 1,
  قيمة_منطقية: true,
  نص: "مرحبا",
  1: "رقم واحد",
  "رقم": 1
};`,
		},
		{
			name:  "should parse a spread element in object",
			input: `متغير شيئ = { رقم: 1, قيمة_منطقية: صحيح, نص: 'مرحبا', ...{ رقم: 1, قيمة_منطقية: صحيح, نص: 'مرحبا' } }؛`,
			expected: `let شيئ = {
  رقم: 1,
  قيمة_منطقية: true,
  نص: "مرحبا",
  ...{
    رقم: 1,
    قيمة_منطقية: true,
    نص: "مرحبا"
  }
};`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(tt.input)
			output := WriteSourceFile(sourceFile)

			if output != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s\n", tt.expected, output)
			}
		})
	}
}

func TestMemberExpression(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple dot access",
			input:    "متغير س = شخص.اسم؛",
			expected: "let س = شخص.اسم;",
		},
		{
			name:     "nested dot access",
			input:    "متغير س = شخص.تفاصيل.عمر؛",
			expected: "let س = شخص.تفاصيل.عمر;",
		},
		// {
		// 	name:     "computed property access",
		// 	input:    "متغير س = شخص['اسم']؛",
		// 	expected: `let س = شخص["اسم"];`,
		// },
		// {
		// 	name:     "nested computed and dot access",
		// 	input:    "متغير س = شخص.تفاصيل['عمر']؛",
		// 	expected: `let س = شخص.تفاصيل["عمر"];`,
		// },
		// {
		// 	name:     "array access",
		// 	input:    "متغير س = ارقام[0]؛",
		// 	expected: "let س = ارقام[0];",
		// },
		// {
		// 	name:     "nested array and object access",
		// 	input:    "متغير س = اشخاص[0].اسم؛",
		// 	expected: "let س = اشخاص[0].اسم;",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(tt.input)
			output := WriteSourceFile(sourceFile)

			if output != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s\n", tt.expected, output)
			}
		})
	}
}

func TestFunctionDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty function declaration",
			input:    "دالة مرحبا() {}",
			expected: "function مرحبا() {}",
		},
		{
			name:     "function with single parameter",
			input:    "دالة اطبع(نص) {}",
			expected: "function اطبع(نص) {}",
		},
		{
			name:     "function with multiple parameters",
			input:    "دالة جمع(أ, ب) {}",
			expected: "function جمع(أ, ب) {}",
		},
		{
			name: "function with a single statement in body",
			input: `دالة مرحبا() {
  اطبع("أهلا")؛
}`,
			expected: `function مرحبا() {
  اطبع("أهلا");
}`,
		},
		{
			name: "function with multiple statements in body",
			input: `دالة عملية(أ, ب) {
  متغير مجموع = أ + ب؛
  متغير طرح = أ - ب؛
}`,
			expected: `function عملية(أ, ب) {
  let مجموع = أ + ب;
  let طرح = أ - ب;
}`,
		},
		{
			name: "nested blocks inside function",
			input: `دالة تحقق(قيمة) {
  إذا(قيمة) {
    اطبع("صحيح")؛
  }
}`,
			expected: `function تحقق(قيمة) {
  if (قيمة) {
    اطبع("صحيح");
  }
}`,
		},
		{
			name:     "function with return statement",
			input:    "دالة جمع(أ, ب) { إرجاع أ + ب؛ }",
			expected: "function جمع(أ, ب) {\n  return أ + ب;\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sourceFile := parser.ParseSourceFile(tt.input)
			output := WriteSourceFile(sourceFile)

			if output != tt.expected {
				t.Errorf("Expected:\n%s\nGot:\n%s\n", tt.expected, output)
			}
		})
	}
}

func TestUpdateExpression(t *testing.T) {
	t.Run("should write a update expression", func(t *testing.T) {
		input := "أ++؛"
		output := WriteSourceFile(parser.ParseSourceFile(input))
		expected := "أ++;"

		if output != expected {
			t.Errorf("\nExpected %s, got %s\n", expected, output)
		}
	})
}

func TestForStatement(t *testing.T) {
	t.Run("should parse a for statement", func(t *testing.T) {
		input := "من_أجل (متغير أ = 0؛ أ <= 10؛ أ++) { أ؛ }"
		output := WriteSourceFile(parser.ParseSourceFile(input))
		expected := "for (let أ = 0; أ <= 10; أ++) {\n  أ;\n}"

		if output != expected {
			t.Errorf("\nExpected %s, got %s\n", expected, output)
		}
	})
}

func TestAssignmentExpression(t *testing.T) {
	t.Run("should parse assignment expression", func(t *testing.T) {
		input := "أ = 100؛"
		output := WriteSourceFile(parser.ParseSourceFile(input))
		expected := "أ = 100;"

		if output != expected {
			t.Errorf("\nExpected %s, got %s\n", expected, output)
		}
	})
}
