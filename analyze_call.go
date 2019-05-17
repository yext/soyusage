package soyusage

import (
	"github.com/robfig/soy/ast"
)

func analyzeCall(
	s *scope,
	call *ast.CallNode,
) error {
	template, found := s.registry.Template(call.Name)
	if !found {
		return newErrorf(s, call, "template not found: %s", call.Name)
	}

	callScope := s.call(call.Name)

	if callScope.callCycles() > s.config.RecursionDepth {
		return nil
	}

	for _, parameter := range call.Params {
		switch v := parameter.(type) {
		case *ast.CallParamContentNode:
			err := analyzeNode(s, UsageFull, v.Content)
			if err != nil {
				return wrapError(s, parameter, err)
			}
			constants, err := extractConstantVariables(s, v.Content)
			if err != nil {
				return wrapError(s, parameter, err)
			}
			callScope.variables[v.Key] = append(callScope.variables[v.Key], constants...)
		case *ast.CallParamValueNode:
			variables, err := extractVariables(s, v.Value)
			if err != nil {
				return wrapError(s, parameter, err)
			}
			callScope.variables[v.Key] = append(callScope.variables[v.Key], variables...)
		}
	}

	if call.AllData {
		for _, templateParam := range template.Doc.Params {
			_, paramPopulated := callScope.parameters[templateParam.Name]
			_, variablePopulated := callScope.variables[templateParam.Name]
			if !paramPopulated && !variablePopulated {
				if paramValue, exists := s.parameters[templateParam.Name]; exists {
					callScope.parameters[templateParam.Name] = paramValue
				}
				if variableValues, exists := s.variables[templateParam.Name]; exists {
					callScope.variables[templateParam.Name] = variableValues
				}
			}
		}
	}

	if call.Data != nil {
		variables, err := extractVariables(s, call.Data)
		if err != nil {
			return wrapError(s, call.Data, err)
		}
		for _, param := range variables {
			for name, param := range param.Children {
				callScope.variables[name] = append(callScope.variables[name], param)
			}
			for _, templateParam := range template.Doc.Params {
				_, paramPopulated := callScope.parameters[templateParam.Name]
				_, variablePopulated := callScope.variables[templateParam.Name]
				if !paramPopulated && !variablePopulated {
					p := newParam()
					if callScope.callCycles() == s.config.RecursionDepth {
						p.addUsageToLeaves(Usage{
							Type:     UsageFull,
							Template: callScope.templateName,
							node:     getNodeForName(s, templateParam.Name, call),
						})
					}
					param.Children[templateParam.Name] = p
					callScope.parameters[templateParam.Name] = p
				}
			}
		}
	}

	if err := analyzeNode(callScope, usageUndefined, template.Node); err != nil {
		return wrapError(s, template.Node, err)
	}

	return nil
}

func getNodeForName(
	s *scope,
	name string,
	call *ast.CallNode,
) ast.Node {
	for _, parameter := range call.Params {
		switch v := parameter.(type) {
		case *ast.CallParamContentNode:
			if v.Key == name {
				return v
			}
		case *ast.CallParamValueNode:
			if v.Key == name {
				return v
			}
		}
	}
	return call
}
