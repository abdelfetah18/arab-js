package transformer

import (
	"arab_js/internal/binder"
	"arab_js/internal/checker"
	"arab_js/internal/compiler/ast"
)

type Program interface {
	SourceFiles() []*ast.SourceFile
	CheckSourceFiles() *binder.NameResolver
}

type Transformer struct {
	program      Program
	NameResolver *binder.NameResolver
	TypeResolver *checker.TypeResolver
}

func NewTransformer(program Program) *Transformer {
	nameResolver := program.CheckSourceFiles()
	return &Transformer{
		program:      program,
		NameResolver: nameResolver,
		TypeResolver: checker.NewTypeResolver(nameResolver),
	}
}

func (t *Transformer) transformNode(node *ast.Node) bool {
	switch node.Type {
	case ast.NodeTypeIdentifier:
		identifier := node.AsIdentifier()
		symbol := t.NameResolver.Resolve(identifier.Name, node)
		if symbol != nil && symbol.OriginalName != nil {
			identifier.Name = *symbol.OriginalName
		}
	case ast.NodeTypeMemberExpression:
		t.transformMemberExpression(node.AsMemberExpression())
	case ast.NodeTypeObjectExpression:
		t.transformObjectExpression(node.AsObjectExpression())
	default:
		node.ForEachChild(func(n *ast.Node) bool { return t.transformNode(n) })
	}
	return false // To Visit All Nodes
}

func (t *Transformer) Transform() {
	for _, sourceFile := range t.program.SourceFiles() {
		if sourceFile.IsDeclarationFile {
			continue
		}

		sourceFile.ForEachChild(t.transformNode)
	}
}

func (t *Transformer) transformMemberExpression(memberExpression *ast.MemberExpression) {
	switch memberExpression.Object.Type {
	case ast.NodeTypeMemberExpression:
		objectType := t.TypeResolver.ResolveTypeFromNode(memberExpression.Object)
		t.transformMemberExpression(memberExpression.Object.AsMemberExpression())
		t.transformProperty(memberExpression.Property, objectType.AsObjectType(), memberExpression.Computed)
	case ast.NodeTypeIdentifier:
		identifier := memberExpression.Object.AsIdentifier()
		symbol := t.NameResolver.Resolve(identifier.Name, identifier.AsNode())
		if symbol != nil && symbol.OriginalName != nil {
			identifier.Name = *symbol.OriginalName
		}

		objectType := t.TypeResolver.ResolveTypeFromNode(symbol.Node)
		if objectType != nil {
			if objectType.Flags&checker.TypeFlagsObject == checker.TypeFlagsObject {
				if objectType.ObjectFlags&checker.ObjectFlagsArrayLiteral != 0 {
					return
				}
				t.transformProperty(memberExpression.Property, objectType.AsObjectType(), memberExpression.Computed)
			} else {
				apparentType := t.TypeResolver.GetApparentType(objectType)
				t.transformProperty(memberExpression.Property, apparentType.AsObjectType(), memberExpression.Computed)
			}
		}
	}
}

func (t *Transformer) transformProperty(property *ast.Node, objectType *checker.ObjectType, isComputed bool) {
	switch property.Type {
	case ast.NodeTypeIdentifier:
		identifier := property.AsIdentifier()
		if !isComputed {
			propertyType := objectType.Members()[identifier.Name]
			if propertyType.OriginalName != nil {
				identifier.Name = *propertyType.OriginalName
			}
		}
	}
}

func (t *Transformer) transformObjectExpression(objectExpression *ast.ObjectExpression) {
	_type := t.TypeResolver.ResolveTypeFromNode(objectExpression.AsNode())
	if _type == nil || _type.Flags&checker.TypeFlagsObject == 0 {
		for _, property := range objectExpression.Properties {
			if property.Type == ast.NodeTypeObjectProperty {
				objectProperty := property.AsObjectProperty()
				t.transformNode(objectProperty.Value)
			}
		}
		return
	}

	objectType := _type.AsObjectType()
	for _, property := range objectExpression.Properties {
		if property.Type == ast.NodeTypeObjectProperty {
			objectProperty := property.AsObjectProperty()
			if objectProperty.Key.Type == ast.NodeTypeIdentifier {
				identifier := objectProperty.Key.AsIdentifier()
				propertyType := objectType.Members()[identifier.Name]
				if propertyType.OriginalName != nil {
					identifier.Name = *propertyType.OriginalName
				}

			}
			t.transformNode(objectProperty.Value)
		}
	}
}
