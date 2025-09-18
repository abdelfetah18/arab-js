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
