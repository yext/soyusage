package soyusage

import (
	"fmt"

	"github.com/robfig/soy/ast"
	"github.com/robfig/soy/template"
)

// AnalyzeTemplate walks the AST for the specified template and outputs a parameter
// tree defining where and how those parameters are used.
func AnalyzeTemplate(name string, registry *template.Registry) (Params, error) {
	template, found := registry.Template(name)
	if !found {
		return nil, fmt.Errorf("template not found: %s", name)
	}

	s := &scope{
		registry:     registry,
		templateName: name,
		parameters:   make(Params),
		variables:    make(map[string][]*Param),
	}

	err := analyzeNode(s, usageUndefined, template.Node)
	if err != nil {
		return nil, err
	}

	// Filter out all the params that are not passed into this template
	var filteredParams = make(Params)
	for _, paramDoc := range template.Doc.Params {
		name := paramDoc.Name
		if param, exists := s.parameters[name]; exists {
			filteredParams[name] = param
		}
	}

	return filteredParams, nil
}

// scope represents the usage at the current position in the stack
type scope struct {
	registry      *template.Registry
	templateName  string
	templateStack []string
	parameters    Params
	variables     map[string][]*Param
}

// inner creates a new scope "inside" the current scope
// The new scope has all the same state, but a new set of variables
// is created so assignments don't escape up the stack.
func (s *scope) inner() *scope {
	out := &scope{
		registry:      s.registry,
		templateName:  s.templateName,
		templateStack: nil,
		parameters:    s.parameters,
		variables:     make(map[string][]*Param),
	}

	for _, template := range s.templateStack {
		out.templateStack = append(out.templateStack, template)
	}

	for name, params := range s.variables {
		out.variables[name] = params
	}
	return out
}

func analyzeNode(s *scope, usageType UsageType, node ...ast.Node) error {
	// Create a new scope for this set of nodes
	cs := s.inner()
	for _, node := range node {
		err := func() error {
			switch v := node.(type) {
			case *ast.AddNode:
				return analyzeNode(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.AndNode:
				return analyzeNode(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.CallNode:
				return analyzeCall(cs, v)
			case *ast.CssNode:
				return analyzeNode(s, UsageFull, v.Children()...)
			case *ast.DataRefNode:
				if _, err := recordDataRef(cs, usageType, v); err != nil {
					return err
				}
			case *ast.DivNode:
				return analyzeNode(cs, UsageFull, v.Children()...)
			case *ast.ElvisNode:
				return analyzeNode(cs, usageType, v.Arg1, v.Arg2)
			case *ast.EqNode:
				return analyzeNode(cs, UsageFull, v.Children()...)
			case *ast.ForNode:
				variables, err := extractVariables(cs, v.List)
				if err != nil {
					return wrapError(s, node, err)
				}
				cs.variables[v.Var] = variables
				constants, err := constantValues(cs, v.List)
				if err != nil {
					return wrapError(s, node, err)
				}
				cs.variables[v.Var] = appendConstants(cs.variables[v.Var], constants...)
				return analyzeNode(cs, usageType, v.Body)
			case *ast.FunctionNode:
				var usage = UsageUnknown
				switch v.Name {
				case "isFirst", "isLast", "index", "isNonnull", "length":
					usage = UsageMeta
				case "keys":
					usage = UsageMeta
				case "augmentMap", "quoteKeysIfJs":
					usage = UsageReference
				case "round", "floor", "ceiling", "min", "max", "randomInt", "strContains":
					usage = UsageFull
				}
				return analyzeNode(cs, usage, v.Children()...)
			case *ast.GlobalNode:
				// Globals assign primitive values and can be ignored for analyzing parameters
			case *ast.GtNode:
				return analyzeNode(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.GteNode:
				return analyzeNode(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.IfNode:
				for _, condition := range v.Conds {
					condUsage := UsageFull
					if _, isDataRef := condition.Cond.(*ast.DataRefNode); isDataRef {
						condUsage = UsageExists
					}
					err := analyzeNode(cs, condUsage, condition.Cond)
					if err != nil {
						return err
					}
					err = analyzeNode(cs, usageType, condition.Body)
					if err != nil {
						return err
					}
				}
				return nil
			case *ast.LetContentNode:
				variables, err := extractConstantVariables(cs, v.Body)
				if err != nil {
					return wrapError(s, node, err)
				}
				cs.variables[v.Name] = variables
			case *ast.LetValueNode:
				variables, err := extractVariables(cs, v.Expr)
				if err != nil {
					return wrapError(s, node, err)
				}
				cs.variables[v.Name] = variables
				return nil
			case *ast.ListLiteralNode:
				return analyzeNode(cs, usageType, v.Items...)
			case *ast.ListNode:
				return analyzeNode(cs, usageType, v.Children()...)
			case *ast.LogNode:
				return analyzeNode(cs, UsageFull, v.Body)
			case *ast.LtNode:
				return analyzeNode(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.LteNode:
				return analyzeNode(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.MapLiteralNode:
				for _, node := range v.Items {
					if err := analyzeNode(cs, usageType, node); err != nil {
						return err
					}
				}
			case *ast.ModNode:
				return analyzeNode(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.MsgNode:
				return analyzeNode(cs, usageType, v.Body)
			case *ast.MsgPlaceholderNode:
				return analyzeNode(cs, usageType, v.Body)
			case *ast.MsgPluralCaseNode:
				return analyzeNode(cs, usageType, v.Body)
			case *ast.MsgPluralNode:
				if err := analyzeNode(cs, UsageFull, v.Value); err != nil {
					return err
				}
				for _, c := range v.Cases {
					if err := analyzeNode(cs, usageType, c.Body); err != nil {
						return err
					}
				}
			case *ast.MulNode:
				return analyzeNode(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.NegateNode:
				return analyzeNode(cs, UsageFull, v.Arg)
			case *ast.NotEqNode:
				return analyzeNode(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.NotNode:
				return analyzeNode(cs, UsageFull, v.Arg)
			case *ast.PrintDirectiveNode:
				return analyzeNode(cs, usageType, v.Children()...)
			case *ast.PrintNode:
				err := analyzeNode(cs, UsageFull, v.Arg)
				if err != nil {
					return err
				}
				for _, directive := range v.Directives {
					err := analyzeNode(cs, UsageUnknown, directive)
					if err != nil {
						return err
					}
				}
			case *ast.SwitchNode:
				if err := analyzeNode(cs, UsageFull, v.Value); err != nil {
					return err
				}
				for _, c := range v.Cases {
					if err := analyzeNode(cs, UsageFull, c.Values...); err != nil {
						return err
					}
					if err := analyzeNode(cs, usageType, c.Body); err != nil {
						return err
					}
				}
			case *ast.TemplateNode:
				return analyzeNode(cs, usageType, v.Children()...)
			case *ast.TernNode:
				return analyzeNode(cs, usageType, v.Arg1, v.Arg2, v.Arg3)
			case *ast.SubNode:
				return analyzeNode(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.OrNode:
				return analyzeNode(cs, UsageFull, v.Arg1, v.Arg2)
			case
				*ast.StringNode,
				*ast.RawTextNode,
				*ast.NullNode,
				*ast.LiteralNode,
				*ast.FloatNode,
				*ast.BoolNode,
				*ast.IntNode,
				*ast.DebuggerNode,
				*ast.IdentNode,
				*ast.MsgHtmlTagNode,
				nil:
			default:
				return fmt.Errorf("unexpected node type: %T", node)
			}
			return nil
		}()
		if err != nil {
			return wrapError(s, node, err)
		}
	}

	return nil
}

func analyzeCall(
	s *scope,
	call *ast.CallNode,
) error {
	var scopes []*scope
	template, found := s.registry.Template(call.Name)
	if !found {
		return newErrorf(s, call, "template not found: %s", call.Name)
	}

	callScope := s.inner()
	if !call.AllData {
		callScope.parameters = make(Params)
	}
	scopes = []*scope{
		callScope,
	}
	if call.Data != nil {
		variables, err := extractVariables(s, call.Data)
		if err != nil {
			return wrapError(s, call.Data, err)
		}
		for _, param := range variables {
			dataScope := s.inner()
			dataScope.variables = make(map[string][]*Param)
			dataScope.parameters = param.Children
			scopes = append(scopes, dataScope)
		}
	}

	for _, scope := range scopes {
		for _, parameter := range call.Params {
			switch v := parameter.(type) {
			case *ast.CallParamContentNode:
				err := analyzeNode(s, UsageFull, v.Content)
				if err != nil {
					return wrapError(s, parameter, err)
				}
			case *ast.CallParamValueNode:
				variables, err := extractVariables(s, v.Value)
				if err != nil {
					return wrapError(s, parameter, err)
				}
				scope.variables[v.Key] = variables
			}
		}
		scope.templateStack = append(scope.templateStack, scope.templateName)
		var cycles int
		for _, stackTemplate := range scope.templateStack {
			if call.Name == stackTemplate {
				cycles++
			}
			if cycles > 5 {
				return nil
			}
		}

		scope.templateName = call.Name
		if err := analyzeNode(scope, usageUndefined, template.Node); err != nil {
			return wrapError(s, template.Node, err)
		}
	}
	return nil
}

func findParams(
	s *scope,
	name string,
) ([]*Param, error) {
	if params, exist := s.variables[name]; exist {
		return params, nil
	}
	if _, exists := s.parameters[name]; !exists {
		s.parameters[name] = newParam()
	}
	return []*Param{
		s.parameters[name],
	}, nil
}

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
			if _, exists := param.Children["[?]"]; !exists {
				param.Children["[?]"] = newParam()
			}
			nextParam = param.Children["[?]"]
		case string:
			if _, exists := param.Children[name]; !exists {
				param.Children[name] = newParam()
			}
			nextParam = param.Children[name]
		}
		leaves, err := recordDataRefAccess(s, usageType, nextParam, access[1:])
		if err != nil {
			return nil, wrapError(s, head, err)
		}
		out = append(out, leaves...)
	}
	return out, nil
}

func constantValues(s *scope, node ast.Node) ([]interface{}, error) {
	switch v := node.(type) {
	case *ast.StringNode:
		return []interface{}{v.Value}, nil
	case *ast.IntNode:
		return []interface{}{int(v.Value)}, nil
	case *ast.DataRefNode:
		params, err := findParams(s, v.Key)
		if err != nil {
			return nil, wrapError(s, v, err)
		}
		var out []interface{}
		for _, param := range params {
			if param.isConstant() {
				out = append(out, param.constant)
			} else {
				out = append(out, nonConstant{})
			}
		}
		return out, nil
	case *ast.AddNode:
		arg1Values, err := constantValues(s, v.Arg1)
		if err != nil {
			return nil, wrapError(s, v, err)
		}
		arg2Values, err := constantValues(s, v.Arg2)
		if err != nil {
			return nil, wrapError(s, v, err)
		}
		var out = make(map[string]struct{})
		for _, arg1 := range arg1Values {
			for _, arg2 := range arg2Values {
				_, arg1IsString := arg1.(string)
				_, arg2IsString := arg2.(string)
				if arg1IsString || arg2IsString {
					out[fmt.Sprint(arg1)+fmt.Sprint(arg2)] = struct{}{}
				}
			}
		}
		return stringSetToInterface(out), nil
	case *ast.FunctionNode:
		if v.Name == "keys" {
			return constantValues(s, v.Args[0])
		}
		if v.Name == "range" {
			var out = make(map[int]struct{})
			var (
				starts     = []interface{}{0}
				ends       []interface{}
				increments = []interface{}{1}
				err        error
			)
			if len(v.Args) == 1 {
				ends, err = constantValues(s, v.Args[0])
				if err != nil {
					return nil, wrapError(s, v, err)
				}
			}
			if len(v.Args) > 1 {
				if starts, err = constantValues(s, v.Args[0]); err != nil {
					return nil, wrapError(s, v, err)
				}
				if ends, err = constantValues(s, v.Args[1]); err != nil {
					return nil, wrapError(s, v, err)
				}
			}
			if len(v.Args) > 2 {
				if increments, err = constantValues(s, v.Args[2]); err != nil {
					return nil, wrapError(s, v, err)
				}
			}
			for _, increment := range increments {
				var (
					incrementI, startI, endI int
					isInt                    bool
				)
				if incrementI, isInt = increment.(int); !isInt {
					continue
				}
				for _, start := range starts {
					for _, end := range ends {
						if startI, isInt = start.(int); !isInt {
							continue
						}
						if endI, isInt = end.(int); !isInt {
							continue
						}
						for i := startI; i < endI; i += incrementI {
							out[i] = struct{}{}
						}
					}
				}
			}
			return intSetToInterface(out), nil
		}
		return []interface{}{nonConstant{}}, nil
	}
	return nil, nil
}

func intSetToInterface(set map[int]struct{}) []interface{} {
	var r []interface{}
	for val := range set {
		r = append(r, val)
	}
	return r
}

func stringSetToInterface(set map[string]struct{}) []interface{} {
	var r []interface{}
	for val := range set {
		r = append(r, val)
	}
	return r
}

func extractConstantVariables(
	s *scope,
	node ast.Node,
) ([]*Param, error) {
	if err := analyzeNode(s, UsageFull, node); err != nil {
		return nil, wrapError(s, node, err)
	}
	if l, isList := node.(*ast.ListNode); isList && len(l.Nodes) == 1 {
		var params []*Param
		switch v := l.Nodes[0].(type) {
		case *ast.RawTextNode:
			p := newParam()
			p.constant = v.String()
			params = append(params, p)
		case *ast.SwitchNode:
			for _, c := range v.Cases {
				p, err := extractConstantVariables(s, c.Body)
				if err != nil {
					return nil, wrapError(s, c, err)
				}
				params = append(params, p...)
			}
		case *ast.IfNode:
			for _, c := range v.Conds {
				p, err := extractConstantVariables(s, c.Body)
				if err != nil {
					return nil, wrapError(s, c, err)
				}
				params = append(params, p...)
			}
		case *ast.MsgPlaceholderNode:
			p, err := extractConstantVariables(s, v.Body)
			if err != nil {
				return nil, wrapError(s, v, err)
			}
			params = append(params, p...)
		case *ast.MsgNode:
			p, err := extractConstantVariables(s, v.Body)
			if err != nil {
				return nil, wrapError(s, v, err)
			}
			params = append(params, p...)
		case *ast.PrintNode:
			if len(v.Directives) == 0 {
				constants, err := constantValues(s, v.Arg)
				if err != nil {
					return nil, wrapError(s, v, err)
				}
				for _, value := range constants {
					p := newParam()
					p.constant = fmt.Sprint(value)
					params = append(params, p)
				}
			}
		case *ast.CallNode, *ast.ForNode:
		default:
			return nil, newErrorf(s, v, "unexpected type: %T\n", v)
		}
		return params, nil
	}
	return nil, nil
}

func extractVariables(
	s *scope,
	node ast.Node,
) ([]*Param, error) {
	var out []*Param
	switch v := node.(type) {
	case *ast.StringNode:
		p := newParam()
		p.constant = v.Value
		out = append(out, p)
	case *ast.IntNode:
		p := newParam()
		p.constant = int(v.Value)
		out = append(out, p)
	case *ast.ListLiteralNode:
		for _, item := range v.Items {
			p, err := extractVariables(s, item)
			if err != nil {
				return nil, wrapError(s, item, err)
			}
			out = append(out, p...)
		}
	case *ast.MapLiteralNode:
		for _, item := range v.Items {
			p, err := extractVariables(s, item)
			if err != nil {
				return nil, wrapError(s, item, err)
			}
			out = append(out, p...)
		}
	case *ast.DataRefNode:
		p, err := recordDataRef(s, UsageReference, v)
		if err != nil {
			return nil, wrapError(s, node, err)
		}
		out = append(out, p...)

		constants, err := constantValues(s, v)
		if err != nil {
			return nil, wrapError(s, v, err)
		}
		out = appendConstants(out, constants...)
	case *ast.ElvisNode:
		v1, err := extractVariables(s, v.Arg1)
		if err != nil {
			return nil, wrapError(s, node, err)
		}
		out = append(out, v1...)
		v2, err := extractVariables(s, v.Arg2)
		if err != nil {
			return nil, wrapError(s, node, err)
		}
		out = append(out, v2...)
	case *ast.TernNode:
		if err := analyzeNode(s, UsageFull, v.Arg1); err != nil {
			return nil, wrapError(s, node, err)
		}
		v1, err := extractVariables(s, v.Arg2)
		if err != nil {
			return nil, wrapError(s, node, err)
		}
		out = append(out, v1...)
		v2, err := extractVariables(s, v.Arg3)
		if err != nil {
			return nil, wrapError(s, node, err)
		}
		out = append(out, v2...)
	case *ast.FunctionNode:
		if v.Name == "augmentMap" || v.Name == "quoteKeysIfJs" {
			for _, arg := range v.Args {
				variables, err := extractVariables(s, arg)
				if err != nil {
					return nil, wrapError(s, node, err)
				}
				out = append(out, variables...)
			}
		}
	default:
		type withChildren interface {
			Children() []ast.Node
		}
		if parent, hasChildren := node.(withChildren); hasChildren {
			if err := analyzeNode(s, UsageFull, parent.Children()...); err != nil {
				return nil, wrapError(s, node, err)
			}
		}

	}
	return out, nil
}

type nonConstant struct{}

func (n nonConstant) String() string {
	return "?"
}

func appendConstants(params []*Param, constants ...interface{}) []*Param {
	out := params
	for _, value := range constants {
		p := newParam()
		p.constant = value
		out = append(out, p)
	}
	return out
}
