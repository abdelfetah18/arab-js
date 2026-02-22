package lexer

import "testing"

func TestComments(t *testing.T) {
	t.Run("detect original name directive", func(t *testing.T) {
		input := "// @الاسم_الأصلي(\"console\")\nاسم"

		lexer := NewLexer(input)
		if !lexer.HasPrecedingOriginalNameDirective || lexer.OriginalNameDirectiveValue != "console" {
			t.Error("Expected to detect original name directive")
		}
	})
}

func TestNumericLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		typ      TokenType
	}{
		{"123", "123", Decimal},
		{"123_456", "123_456", Decimal},
		{"0.5", "0.5", Decimal},
		{"1.2e3", "1.2e3", Decimal},
		{"0o123", "0o123", Decimal},
		{"0O123", "0O123", Decimal},
		{"0o123_456", "0o123_456", Decimal},
		{"0o777", "0o777", Decimal},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			lexer := NewLexer(tt.input)
			token := lexer.Peek()
			if token.Type != tt.typ {
				t.Errorf("Expected token type %v, got %v", tt.typ, token.Type)
			}
			if token.Value != tt.expected {
				t.Errorf("Expected token value %v, got %v", tt.expected, token.Value)
			}
		})
	}
}

func TestInvalidNumericLiterals(t *testing.T) {
	tests := []string{
		"0o",
		"0o_",
		"0o123_",
		"123_",
		"0o8", // Not an octal digit
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			lexer := NewLexer(input)
			token := lexer.Peek()
			// If it's 0o8, it might lex 0o then fail on 8, or partially lex.
			// Actually 0o8 should fail to lex as a single octal literal and return 0o as invalid and then 8.
			// In our implementation, 0o8 will lex 0 then o then fail because 8 is not octal digit.
			// So it returns Invalid for '0'.
			if token.Type == Decimal && token.Value == input {
				t.Errorf("Expected invalid literal for %v, but got valid Decimal %v", input, token.Value)
			}
		})
	}
}
