package binder

import (
	"arab_js/internal/compiler"
	"arab_js/internal/compiler/ast"
	"testing"
)

func TestBindVariableDeclaration(t *testing.T) {
	t.Run("should bind variable declaration at global scope", func(t *testing.T) {
		input := "متغير رقم : عدد = 100؛"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("رقم", sourceFile.Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in global scope, but got nil", "رقم")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsNumber {
			t.Errorf("expected symbol %q to have type %v, but got %v", "رقم", ast.TypeFlagsNumber, symbol.Type.Flags)
			return
		}

	})

	t.Run("should bind variable declaration at block scope", func(t *testing.T) {
		input := "{ متغير رقم : عدد = 100؛ }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("رقم", sourceFile.Body[0].AsBlockStatement().Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in block scope, but got nil", "رقم")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsNumber {
			t.Errorf("expected symbol %q to have type %v, but got %v", "رقم", ast.TypeFlagsNumber, symbol.Type.Flags)
			return
		}
	})

	t.Run("should bind variable declaration at function scope", func(t *testing.T) {
		input := "دالة تجربة() { متغير رقم : عدد = 100؛ }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("رقم", sourceFile.Body[0].AsFunctionDeclaration().Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in function scope, but got nil", "رقم")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsNumber {
			t.Errorf("expected symbol %q to have type %v, but got %v", "رقم", ast.TypeFlagsNumber, symbol.Type.Flags)
			return
		}
	})

	t.Run("should bind variable declaration at for statement init", func(t *testing.T) {
		input := "من_أجل (متغير رقم : عدد = 0؛ رقم <= 10؛ رقم++) {}"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("رقم", sourceFile.Body[0].AsForStatement().Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist, but got nil", "رقم")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsNumber {
			t.Errorf("expected symbol %q to have type %v, but got %v", "رقم", ast.TypeFlagsNumber, symbol.Type.Flags)
			return
		}
	})

	t.Run("should bind variable declaration at for statement scope", func(t *testing.T) {
		input := "من_أجل (متغير أ : عدد = 0؛ أ <= 10؛ أ++) { متغير رقم: عدد = 100؛ }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("رقم", sourceFile.Body[0].AsForStatement().Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist, but got nil", "رقم")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsNumber {
			t.Errorf("expected symbol %q to have type %v, but got %v", "رقم", ast.TypeFlagsNumber, symbol.Type.Flags)
			return
		}
	})
}

func TestBindFunctionDeclaration(t *testing.T) {
	t.Run("should bind function declaration at global scope", func(t *testing.T) {
		input := "دالة تجربة() { متغير رقم : عدد = 100؛ }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("تجربة", sourceFile.Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in global scope, but got nil", "تجربة")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsFunction {
			t.Errorf("expected symbol %q to have type %v, but got %v", "تجربة", ast.TypeFlagsFunction, symbol.Type.Flags)
			return
		}
	})

	t.Run("should bind function declaration at block scope", func(t *testing.T) {
		input := "{ دالة تجربة() { متغير رقم : عدد = 100؛ } }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("تجربة", sourceFile.Body[0].AsBlockStatement().Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in block scope, but got nil", "تجربة")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsFunction {
			t.Errorf("expected symbol %q to have type %v, but got %v", "تجربة", ast.TypeFlagsFunction, symbol.Type.Flags)
			return
		}
	})

	t.Run("should bind function declaration at function scope", func(t *testing.T) {
		input := "دالة خارجية() { دالة تجربة() { متغير رقم : عدد = 100؛ } }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("تجربة", sourceFile.Body[0].AsFunctionDeclaration().Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in function scope, but got nil", "تجربة")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsFunction {
			t.Errorf("expected symbol %q to have type %v, but got %v", "تجربة", ast.TypeFlagsFunction, symbol.Type.Flags)
			return
		}
	})

	t.Run("should bind function declaration at for statement scope", func(t *testing.T) {
		input := "من_أجل (متغير أ : عدد = 0؛ أ <= 10؛ أ++) { دالة تجربة() { متغير رقم : عدد = 100؛ } }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("تجربة", sourceFile.Body[0].AsForStatement().Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in block scope, but got nil", "تجربة")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsFunction {
			t.Errorf("expected symbol %q to have type %v, but got %v", "تجربة", ast.TypeFlagsFunction, symbol.Type.Flags)
			return
		}
	})

	t.Run("should bind function declaration params at function scope", func(t *testing.T) {
		input := "دالة جمع (أ: عدد, ب: عدد) { إرجاع أ + ب؛ }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("أ", sourceFile.Body[0].AsFunctionDeclaration().Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in block scope, but got nil", "أ")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsNumber {
			t.Errorf("expected symbol %q to have type %v, but got %v", "أ", ast.TypeFlagsNumber, symbol.Type.Flags)
			return
		}
	})
}

func TestBindInterfaceDeclaration(t *testing.T) {
	t.Run("should bind interface declaration at global scope", func(t *testing.T) {
		input := "واجهة شخص { اسم: نص }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("شخص", sourceFile.Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in global scope, but got nil", "شخص")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsObject {
			t.Errorf("expected symbol %q to have type %v, but got %v", "تجربة", ast.TypeFlagsObject, symbol.Type.Flags)
			return
		}
	})

	t.Run("should bind interface declaration at block scope", func(t *testing.T) {
		input := "{ واجهة شخص { اسم: نص } }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("شخص", sourceFile.Body[0].AsBlockStatement().Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in block scope, but got nil", "شخص")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsObject {
			t.Errorf("expected symbol %q to have type %v, but got %v", "شخص", ast.TypeFlagsObject, symbol.Type.Flags)
			return
		}
	})

	t.Run("should bind interface declaration at function scope", func(t *testing.T) {
		input := "دالة خارجية() { واجهة شخص { اسم: نص } }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("شخص", sourceFile.Body[0].AsFunctionDeclaration().Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in function scope, but got nil", "شخص")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsObject {
			t.Errorf("expected symbol %q to have type %v, but got %v", "شخص", ast.TypeFlagsObject, symbol.Type.Flags)
			return
		}
	})

	t.Run("should bind interface declaration at for statement scope", func(t *testing.T) {
		input := "من_أجل (متغير أ : عدد = 0؛ أ <= 10؛ أ++) { واجهة شخص { اسم: نص } }"
		sourceFile := compiler.ParseSourceFile(input)
		BindSourceFile(sourceFile)
		nameResolver := NewNameResolver(sourceFile.Scope)
		symbol := nameResolver.Resolve("شخص", sourceFile.Body[0].AsForStatement().Scope)

		if symbol == nil {
			t.Errorf("expected symbol %q to exist in block scope, but got nil", "شخص")
			return
		}

		if symbol.Type.Flags != ast.TypeFlagsObject {
			t.Errorf("expected symbol %q to have type %v, but got %v", "شخص", ast.TypeFlagsObject, symbol.Type.Flags)
			return
		}
	})
}
