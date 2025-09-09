package ast

type NodeVisitor struct {
	Visit func(node *Node) *Node
}

func NewNodeVisitor(visit func(node *Node) *Node) *NodeVisitor {
	return &NodeVisitor{Visit: visit}
}

func (v *NodeVisitor) VisitNode(node *Node) *Node {
	if node == nil || v.Visit == nil {
		return node
	}

	return v.Visit(node)
}

func (v *NodeVisitor) visitProgram(sourceFile *SourceFile) {
	for _, node := range sourceFile.Body {
		v.VisitNode(node)
	}
}

func (v *NodeVisitor) visitBlockStatement(blockStatement *BlockStatement) {
	for _, node := range blockStatement.Body {
		v.VisitNode(node)
	}
}

func (v *NodeVisitor) visitVariableDeclaration(variableDeclaration *VariableDeclaration) {
	v.VisitNode(variableDeclaration.Identifier.AsNode())

	if variableDeclaration.Initializer != nil {
		v.VisitNode(variableDeclaration.Initializer.Expression)
	}
}
