package soyusage

import "github.com/robfig/soy/ast"

func recordDataRef(
	s *scope,
	usageType UsageType,
	node *ast.DataRefNode,
) ([]*Param, error) {
	if usageType == usageUndefined {
		return nil, newErrorf(s, node, "usage type was not set")
	}

	params, err := findParams(s, node.Key)
	if err != nil {
		return nil, wrapError(s, node, err)
	}

	var out []*Param

	for _, param := range params {
		// Skip constants
		if param.isConstant() {
			continue
		}
		leaves, err := recordDataRefAccess(s, usageType, param, node.Access)
		if err != nil {
			return nil, wrapError(s, node, err)
		}

		for _, leaf := range leaves {
			templateUsage := leaf.Usage[s.templateName]
			leaf.Usage[s.templateName] = append(templateUsage, Usage{
				Template: s.templateName,
				Type:     usageType,
				node:     node,
			})
		}
		out = append(out, leaves...)
	}

	return out, nil
}

func recordDataRefAccess(s *scope,
	usageType UsageType,
	param *Param,
	access []ast.Node) ([]*Param, error) {
	if len(access) == 0 {
		return []*Param{param}, nil
	}

	head := access[0]
	var names []interface{}
	switch access := head.(type) {
	case *ast.DataRefKeyNode:
		names = []interface{}{access.Key}
	case *ast.DataRefIndexNode:
		names = []interface{}{access.Index}
	case *ast.DataRefExprNode:
		constantValues, err := constantValues(s, access.Arg)
		if err != nil {
			return nil, wrapError(s, access, err)
		}
		names = append(names, constantValues...)
		err = analyzeNode(s, UsageFull, access.Arg)
		if err != nil {
			return nil, wrapError(s, access, err)
		}
	}
	var out []*Param
	for _, n := range names {
		var nextParam *Param
		switch name := n.(type) {
		case int:
			nextParam = param
		case nonConstant:
			nextParam = param.getChildOrNew("[?]")
		case string:
			nextParam = param.getChildOrNew(name)
		}
		leaves, err := recordDataRefAccess(s, usageType, nextParam, access[1:])
		if err != nil {
			return nil, wrapError(s, head, err)
		}
		out = append(out, leaves...)
	}
	return out, nil
}
