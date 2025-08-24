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
