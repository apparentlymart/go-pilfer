package pilfer

import (
	"go/ast"
)

type astVisitor func(node ast.Node)

func (v astVisitor) Visit(node ast.Node) ast.Visitor {
	if node != nil {
		v(node)
		return v
	}
	return nil
}

func (v astVisitor) VisitAll(node ast.Node) {
	ast.Walk(v, node)
}
