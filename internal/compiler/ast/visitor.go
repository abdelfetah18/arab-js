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

	result := v.Visit(node)

	if node.Type == NodeTypeProgram {
		v.visitProgram(node.AsProgram())
	}

	if node.Type == NodeTypeBlockStatement {
		v.visitBlockStatement(node.AsBlockStatement())
	}

	return result
}

func (v *NodeVisitor) visitProgram(program *Program) {
	for _, node := range program.Body {
		v.VisitNode(node)
	}
}

func (v *NodeVisitor) visitBlockStatement(blockStatement *BlockStatement) {
	for _, node := range blockStatement.Body {
		v.VisitNode(node)
	}
}
