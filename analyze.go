package soyusage

import (
	"errors"
	"fmt"

	"github.com/robfig/soy/ast"
	"github.com/robfig/soy/template"
)

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
		leafUsage:    UsageFull,
	}

	err := analyzeNode(s, template.Node)
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
	registry     *template.Registry
	templateName string
	parameters   Params
	variables    map[string][]*Param
	leafUsage    Usage
}

// inner creates a new scope "inside" the current scope
// The new scope has all the same state, but a new set of variables
// is created so assignments don't escape up the stack.
func (s *scope) inner() *scope {
	out := &scope{
		registry:     s.registry,
		templateName: s.templateName,
		parameters:   s.parameters,
		leafUsage:    s.leafUsage,
		variables:    make(map[string][]*Param),
	}

	for name, params := range s.variables {
		out.variables[name] = params
	}
	return out
}

func analyzeNodeSetUsage(s *scope, usage Usage, node ...ast.Node) error {
	cs := s.inner()
	cs.leafUsage = usage
	return analyzeNode(cs, node...)
}

func analyzeNode(s *scope, node ...ast.Node) error {
	// Create a new scope for this set of nodes
	cs := s.inner()
	for _, node := range node {
		err := func() error {
			switch v := node.(type) {
			case *ast.CallNode:
				return analyzeCall(cs, v)
			case *ast.CssNode:
				return analyzeNode(s, v.Children()...)
			case *ast.DataRefNode:
				_, err := recordDataRef(cs, v)
				if err != nil {
					return err
				}
			case *ast.DivNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Children()...)
			case *ast.ElvisNode:
				return analyzeNode(cs, v.Arg1, v.Arg2)
			case *ast.EqNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Children()...)
			case *ast.ForNode:
				return errors.New("not implemented")
			case *ast.FunctionNode:
				return analyzeNodeSetUsage(cs, UsageUnknown, v.Children()...)
			case *ast.GlobalNode:
				return errors.New("not implemented")
			case *ast.GtNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.GteNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.IfNode:
				for _, condition := range v.Conds {
					err := analyzeNodeSetUsage(cs, UsageFull, condition.Cond)
					if err != nil {
						return err
					}
					err = analyzeNode(s, condition.Body)
					if err != nil {
						return err
					}
				}
				return nil
			case *ast.LetContentNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Body)
			case *ast.LetValueNode:
				// Clear any existing variable value
				cs.variables[v.Name] = nil
				return mapVariable(cs, v.Name, v.Expr)
			case *ast.ListLiteralNode:
				return errors.New("not implemented")
			case *ast.ListNode:
				return analyzeNode(cs, v.Children()...)
			case *ast.LogNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Body)
			case *ast.LtNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.LteNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.MapLiteralNode:
				return errors.New("not implemented")
			case *ast.ModNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.MsgHtmlTagNode:
				return errors.New("not implemented")
			case *ast.MsgNode:
				return errors.New("not implemented")
			case *ast.MsgPlaceholderNode:
				return errors.New("not implemented")
			case *ast.MsgPluralCaseNode:
				return errors.New("not implemented")
			case *ast.MsgPluralNode:
				return errors.New("not implemented")
			case *ast.MulNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.NegateNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Arg)
			case *ast.NotEqNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.NotNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Arg)
			case *ast.PrintDirectiveNode:
				return analyzeNode(cs, v.Children()...)
			case *ast.PrintNode:
				err := analyzeNodeSetUsage(cs, UsageFull, v.Arg)
				if err != nil {
					return err
				}
				for _, directive := range v.Directives {
					err := analyzeNodeSetUsage(cs, UsageUnknown, directive)
					if err != nil {
						return err
					}
				}
			case *ast.SwitchNode:
				return errors.New("not implemented")
			case *ast.TemplateNode:
				return analyzeNode(cs, v.Children()...)
			case *ast.TernNode:
				return analyzeNode(cs, v.Arg1, v.Arg2, v.Arg3)
			case *ast.SubNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Arg1, v.Arg2)
			case *ast.OrNode:
				return analyzeNodeSetUsage(cs, UsageFull, v.Arg1, v.Arg2)
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
				nil:
			default:
				return fmt.Errorf("unexpected node type: %T", node)
			}
			return nil
		}()
		if err != nil {
			return err
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
		return fmt.Errorf("template not found: %s", call.Name)
	}

	if !call.AllData {
		callScope := s.inner()
		callScope.parameters = make(Params)
		scopes = []*scope{
			callScope,
		}
	}
	if call.Data != nil {
		variables, err := extractVariables(s, "data", call.Data)
		if err != nil {
			return err
		}
		for _, param := range variables["data"] {
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
				err := analyzeNodeSetUsage(s, UsageFull, v.Content)
				if err != nil {
					return err
				}
			case *ast.CallParamValueNode:
				variables, err := extractVariables(s, v.Key, v.Value)
				if err != nil {
					return err
				}
				for key, params := range variables {
					scope.variables[key] = params
				}
			}
		}
		scope.templateName = call.Name
		err := analyzeNode(scope, template.Node)
		if err != nil {
			return err
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
	node *ast.DataRefNode,
) ([]*Param, error) {
	params, err := findParams(s, node.Key)
	if err != nil {
		return nil, err
	}

	var out []*Param
	for _, param := range params {
		for _, accessNode := range node.Access {
			switch access := accessNode.(type) {
			case *ast.DataRefKeyNode:
				if _, exists := param.Children[access.Key]; !exists {
					param.Children[access.Key] = newParam()
				}
				param = param.Children[access.Key]
			case *ast.DataRefIndexNode:
			case *ast.DataRefExprNode:
			}
		}
		templateUsage := param.Usage[s.templateName]
		param.Usage[s.templateName] = append(templateUsage, s.leafUsage)
		out = append(out, param)
	}
	return out, nil
}

func mapVariable(
	s *scope,
	name string,
	node ast.Node,
) error {
	variables, err := extractVariables(s, name, node)
	if err != nil {
		return err
	}
	for key, params := range variables {
		s.variables[key] = append(s.variables[key], params...)
	}
	return nil
}

func extractVariables(
	s *scope,
	name string,
	node ast.Node,
) (map[string][]*Param, error) {
	var out = make(map[string][]*Param)
	switch v := node.(type) {
	case *ast.DataRefNode:
		rs := s.inner()
		rs.leafUsage = UsageReference
		p, err := recordDataRef(rs, v)
		if err != nil {
			return nil, err
		}
		out[name] = append(out[name], p...)
	case *ast.ElvisNode:
		v1, err := extractVariables(s, name, v.Arg1)
		if err != nil {
			return nil, err
		}
		for key, params := range v1 {
			out[key] = append(out[key], params...)
		}
		v2, err := extractVariables(s, name, v.Arg2)
		if err != nil {
			return nil, err
		}
		for key, params := range v2 {
			out[key] = append(out[key], params...)
		}
	case *ast.TernNode:
		if err := analyzeNode(s, v.Arg1); err != nil {
			return nil, err
		}
		v1, err := extractVariables(s, name, v.Arg2)
		if err != nil {
			return nil, err
		}
		for key, params := range v1 {
			out[key] = append(out[key], params...)
		}
		v2, err := extractVariables(s, name, v.Arg3)
		if err != nil {
			return nil, err
		}
		for key, params := range v2 {
			out[key] = append(out[key], params...)
		}
	default:
		type withChildren interface {
			Children() []ast.Node
		}
		if parent, hasChildren := node.(withChildren); hasChildren {
			if err := analyzeNode(s, parent.Children()...); err != nil {
				return nil, err
			}
		}

	}
	return out, nil
}
